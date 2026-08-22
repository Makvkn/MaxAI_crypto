package ai

import (
	"context"

	"github.com/maxaicrypto/backend/internal/domain/conversation"
	"github.com/maxaicrypto/backend/internal/domain/usage"
)

// EventType is an SSE event name from the streaming contract (§80, §81).
type EventType string

const (
	EventToolStarted   EventType = "tool_started"
	EventToolCompleted EventType = "tool_completed"
	EventTextDelta     EventType = "text_delta"
	EventCompleted     EventType = "completed"
	EventError         EventType = "error"
)

// Event is one item of an AI response stream. It is an application-level
// abstraction: the HTTP layer serializes it, and no orchestration logic lives
// inside the handler (§82).
type Event struct {
	Type EventType

	// ToolCallID and Tool are set on tool events.
	ToolCallID string
	Tool       conversation.ToolName
	// ToolOK reports whether a completed tool succeeded.
	ToolOK bool

	// Text is the incremental answer fragment on a text_delta event.
	Text string

	// Message is the persisted assistant message on a completed event.
	Message *conversation.Message

	// Usage is today's quota after the operation settles.
	Usage *usage.Daily

	// ErrorCode and ErrorMessage carry a domain-level failure. Provider errors
	// are mapped before reaching this point (§28).
	ErrorCode    string
	ErrorMessage string
}

// StreamSink receives events as the answer is produced. Implementations must
// respect context cancellation so a client disconnect aborts the upstream
// model request (§82).
type StreamSink interface {
	Send(ctx context.Context, event Event) error
}
