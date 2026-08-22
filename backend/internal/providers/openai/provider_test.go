package openai

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/domain/provider"
)

func TestGenerateMapsContentAndUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Fatalf("missing bearer auth")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
		  "model":"gpt-4o-mini",
		  "choices":[{"message":{"content":"Portfolio looks healthy.","tool_calls":[]}}],
		  "usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}
		}`))
	}))
	t.Cleanup(server.Close)

	p := New(config.AIConfig{
		Model:          "gpt-4o-mini",
		APIKey:         "test-key",
		BaseURL:        server.URL,
		RequestTimeout: 5 * time.Second,
	}, config.ProviderConfig{MaxAttempts: 1})

	resp, err := p.Generate(context.Background(), provider.LLMRequest{
		Messages: []provider.LLMMessage{{Role: provider.RoleUser, Content: "Summarize"}},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if resp.Content != "Portfolio looks healthy." {
		t.Fatalf("content = %q", resp.Content)
	}
	if resp.Usage.TotalTokens != 15 {
		t.Fatalf("usage = %+v", resp.Usage)
	}
}

func TestStreamEmitsDeltasAndCompleted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		_, _ = io.WriteString(w, "data: {\"model\":\"gpt-4o-mini\",\"choices\":[{\"delta\":{\"content\":\"Hello\"}}]}\n\n")
		if flusher != nil {
			flusher.Flush()
		}
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\" world\"}}],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":2,\"total_tokens\":3}}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
		if flusher != nil {
			flusher.Flush()
		}
	}))
	t.Cleanup(server.Close)

	p := New(config.AIConfig{
		Model:          "gpt-4o-mini",
		APIKey:         "test-key",
		BaseURL:        server.URL,
		RequestTimeout: 5 * time.Second,
	}, config.ProviderConfig{MaxAttempts: 1})

	sink := &collectingSink{}
	if err := p.Stream(context.Background(), provider.LLMRequest{
		Messages: []provider.LLMMessage{{Role: provider.RoleUser, Content: "Hi"}},
	}, sink); err != nil {
		t.Fatalf("Stream: %v", err)
	}
	if sink.text != "Hello world" {
		t.Fatalf("text = %q", sink.text)
	}
	if sink.completed == nil || sink.completed.Content != "Hello world" {
		t.Fatalf("completed = %+v", sink.completed)
	}
}

type collectingSink struct {
	text      string
	completed *provider.LLMResponse
}

func (s *collectingSink) OnTextDelta(_ context.Context, delta string) error {
	s.text += delta
	return nil
}

func (s *collectingSink) OnToolCall(context.Context, provider.ToolInvocation) error { return nil }

func (s *collectingSink) OnCompleted(_ context.Context, response provider.LLMResponse) error {
	s.completed = &response
	return nil
}
