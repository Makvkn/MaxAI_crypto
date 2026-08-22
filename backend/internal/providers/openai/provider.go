// Package openai adapts OpenAI to the LLM port. The model is configuration, so
// switching vendors is an adapter change rather than a domain change (§67).
package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/provider"
	"github.com/maxaicrypto/backend/internal/providers/httpx"
)

// Provider implements provider.LLMProvider.
type Provider struct {
	client *httpx.Client
	model  string
}

// New builds the adapter from AI configuration.
func New(cfg config.AIConfig, providerCfg config.ProviderConfig) *Provider {
	return &Provider{
		client: httpx.New(provider.OpenAI, httpx.Config{
			BaseURL:         cfg.BaseURL,
			APIKey:          cfg.APIKey,
			Timeout:         cfg.RequestTimeout,
			MaxAttempts:     providerCfg.MaxAttempts,
			BackoffSchedule: providerCfg.BackoffSchedule,
		}),
		model: cfg.Model,
	}
}

// Name implements provider.LLMProvider.
func (p *Provider) Name() provider.Name { return provider.OpenAI }

// Model returns the configured model identifier.
func (p *Provider) Model() string { return p.model }

// Generate implements provider.LLMProvider.
func (p *Provider) Generate(ctx context.Context, req provider.LLMRequest) (provider.LLMResponse, error) {
	payload := p.buildRequest(req, false)
	var raw chatCompletionResponse
	if err := p.postJSON(ctx, "/chat/completions", payload, &raw); err != nil {
		return provider.LLMResponse{}, err
	}
	return mapResponse(raw), nil
}

// Stream implements provider.LLMProvider.
func (p *Provider) Stream(ctx context.Context, req provider.LLMRequest, sink provider.StreamSink) error {
	payload := p.buildRequest(req, true)
	body, err := json.Marshal(payload)
	if err != nil {
		return apperr.Wrap(apperr.CodeProviderError, err).WithDetail("provider", string(provider.OpenAI))
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint("/chat/completions"), bytes.NewReader(body))
	if err != nil {
		return apperr.Wrap(apperr.CodeProviderError, err).WithDetail("provider", string(provider.OpenAI))
	}
	p.setHeaders(httpReq)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.Do(ctx, httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := httpx.MapStatus(provider.OpenAI, resp.StatusCode); err != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return err
	}

	return p.consumeStream(ctx, resp.Body, sink)
}

func (p *Provider) buildRequest(req provider.LLMRequest, stream bool) chatCompletionRequest {
	model := req.Model
	if model == "" {
		model = p.model
	}
	out := chatCompletionRequest{
		Model:    model,
		Messages: mapMessages(req.Messages),
		Stream:   stream,
	}
	if req.MaxTokens > 0 {
		out.MaxTokens = req.MaxTokens
	}
	if req.Temperature != nil {
		out.Temperature = req.Temperature
	}
	if len(req.Tools) > 0 {
		out.Tools = mapTools(req.Tools)
	}
	if len(req.ResponseSchema) > 0 {
		out.ResponseFormat = &responseFormat{
			Type: "json_schema",
			JSONSchema: &jsonSchemaFormat{
				Name:   "ai_response",
				Strict: true,
				Schema: req.ResponseSchema,
			},
		}
	}
	return out
}

func (p *Provider) consumeStream(ctx context.Context, body io.Reader, sink provider.StreamSink) error {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var (
		content   strings.Builder
		toolCalls = map[int]*toolCallAccumulator{}
		usage     provider.TokenUsage
		model     string
	)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk chatCompletionChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if chunk.Model != "" {
			model = chunk.Model
		}
		if chunk.Usage != nil {
			usage = provider.TokenUsage{
				PromptTokens:     chunk.Usage.PromptTokens,
				CompletionTokens: chunk.Usage.CompletionTokens,
				TotalTokens:      chunk.Usage.TotalTokens,
			}
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				content.WriteString(choice.Delta.Content)
				if err := sink.OnTextDelta(ctx, choice.Delta.Content); err != nil {
					return err
				}
			}
			for _, call := range choice.Delta.ToolCalls {
				acc, ok := toolCalls[call.Index]
				if !ok {
					acc = &toolCallAccumulator{}
					toolCalls[call.Index] = acc
				}
				if call.ID != "" {
					acc.ID = call.ID
				}
				if call.Function.Name != "" {
					acc.Name = call.Function.Name
				}
				acc.Arguments.WriteString(call.Function.Arguments)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return apperr.Wrap(apperr.CodeAIStreamFailed, err).WithDetail("provider", string(provider.OpenAI))
	}

	invocations := make([]provider.ToolInvocation, 0, len(toolCalls))
	for i := 0; i < len(toolCalls); i++ {
		acc, ok := toolCalls[i]
		if !ok {
			continue
		}
		invocation := provider.ToolInvocation{
			ID:        acc.ID,
			Name:      acc.Name,
			Arguments: []byte(acc.Arguments.String()),
		}
		invocations = append(invocations, invocation)
		if err := sink.OnToolCall(ctx, invocation); err != nil {
			return err
		}
	}

	return sink.OnCompleted(ctx, provider.LLMResponse{
		Content:   content.String(),
		ToolCalls: invocations,
		Usage:     usage,
		Model:     model,
	})
}

func (p *Provider) postJSON(ctx context.Context, path string, payload any, dst any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return apperr.Wrap(apperr.CodeProviderError, err).WithDetail("provider", string(provider.OpenAI))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint(path), bytes.NewReader(body))
	if err != nil {
		return apperr.Wrap(apperr.CodeProviderError, err).WithDetail("provider", string(provider.OpenAI))
	}
	p.setHeaders(req)

	resp, err := p.client.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := httpx.MapStatus(provider.OpenAI, resp.StatusCode); err != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return err
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return apperr.Wrap(apperr.CodeProviderError, err).WithDetail("provider", string(provider.OpenAI))
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return apperr.Wrap(apperr.CodeProviderError, fmt.Errorf("decode openai response: %w", err)).
			WithDetail("provider", string(provider.OpenAI))
	}
	return nil
}

func (p *Provider) endpoint(path string) string {
	return strings.TrimRight(p.client.BaseURL(), "/") + path
}

func (p *Provider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if key := p.client.APIKey(); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
}

func mapMessages(messages []provider.LLMMessage) []chatMessage {
	out := make([]chatMessage, 0, len(messages))
	for _, message := range messages {
		item := chatMessage{
			Role:    string(message.Role),
			Content: message.Content,
		}
		if message.ToolCallID != nil {
			item.ToolCallID = *message.ToolCallID
		}
		if message.Name != nil {
			item.Name = *message.Name
		}
		out = append(out, item)
	}
	return out
}

func mapTools(tools []provider.ToolDefinition) []chatTool {
	out := make([]chatTool, 0, len(tools))
	for _, tool := range tools {
		schema := tool.ParametersSchema
		if schema == nil {
			schema = map[string]any{"type": "object", "properties": map[string]any{}}
		}
		out = append(out, chatTool{
			Type: "function",
			Function: chatToolFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  schema,
			},
		})
	}
	return out
}

func mapResponse(raw chatCompletionResponse) provider.LLMResponse {
	response := provider.LLMResponse{
		Model: raw.Model,
		Usage: provider.TokenUsage{
			PromptTokens:     raw.Usage.PromptTokens,
			CompletionTokens: raw.Usage.CompletionTokens,
			TotalTokens:      raw.Usage.TotalTokens,
		},
	}
	if len(raw.Choices) == 0 {
		return response
	}
	message := raw.Choices[0].Message
	response.Content = message.Content
	if len(message.ToolCalls) > 0 {
		response.ToolCalls = make([]provider.ToolInvocation, 0, len(message.ToolCalls))
		for _, call := range message.ToolCalls {
			response.ToolCalls = append(response.ToolCalls, provider.ToolInvocation{
				ID:        call.ID,
				Name:      call.Function.Name,
				Arguments: []byte(call.Function.Arguments),
			})
		}
	}
	return response
}

var _ provider.LLMProvider = (*Provider)(nil)
