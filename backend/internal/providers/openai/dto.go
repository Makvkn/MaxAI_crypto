package openai

import "strings"

type chatCompletionRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	Tools          []chatTool      `json:"tools,omitempty"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	Temperature    *float64        `json:"temperature,omitempty"`
	Stream         bool            `json:"stream"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role       string `json:"role"`
	Content    string `json:"content,omitempty"`
	Name       string `json:"name,omitempty"`
	ToolCallID string `json:"tool_call_id,omitempty"`
}

type chatTool struct {
	Type     string           `json:"type"`
	Function chatToolFunction `json:"function"`
}

type chatToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters"`
}

type responseFormat struct {
	Type       string            `json:"type"`
	JSONSchema *jsonSchemaFormat `json:"json_schema,omitempty"`
}

type jsonSchemaFormat struct {
	Name   string         `json:"name"`
	Strict bool           `json:"strict"`
	Schema map[string]any `json:"schema"`
}

type chatCompletionResponse struct {
	Model   string                 `json:"model"`
	Choices []chatCompletionChoice `json:"choices"`
	Usage   chatUsage              `json:"usage"`
}

type chatCompletionChoice struct {
	Message chatResponseMessage `json:"message"`
}

type chatResponseMessage struct {
	Content   string          `json:"content"`
	ToolCalls []chatToolCall  `json:"tool_calls"`
}

type chatToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function chatToolCallFunc   `json:"function"`
	Index    int                `json:"index"`
}

type chatToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type chatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type chatCompletionChunk struct {
	Model   string              `json:"model"`
	Choices []chatChunkChoice   `json:"choices"`
	Usage   *chatUsage          `json:"usage"`
}

type chatChunkChoice struct {
	Delta chatChunkDelta `json:"delta"`
}

type chatChunkDelta struct {
	Content   string         `json:"content"`
	ToolCalls []chatToolCall `json:"tool_calls"`
}

type toolCallAccumulator struct {
	ID        string
	Name      string
	Arguments strings.Builder
}
