package provider

import "context"

// MessageRole separates trusted instructions from user input inside an LLM
// request. External blockchain strings never enter as instructions (§77, §152).
type MessageRole string

const (
	// RoleSystem carries backend-owned policy. It is never assembled from
	// external data.
	RoleSystem MessageRole = "system"
	// RoleUser carries the end user's question.
	RoleUser MessageRole = "user"
	// RoleAssistant carries prior model turns.
	RoleAssistant MessageRole = "assistant"
	// RoleTool carries the structured result of a backend tool execution.
	RoleTool MessageRole = "tool"
)

// LLMMessage is one turn of an LLM request.
type LLMMessage struct {
	Role    MessageRole
	Content string
	// ToolCallID links a RoleTool message to the call it answers.
	ToolCallID *string
	// Name is the tool name for RoleTool messages.
	Name *string
}

// ToolDefinition describes a domain tool the model may request. The backend
// executes the tool; the model never computes financial values itself (§50).
type ToolDefinition struct {
	Name        string
	Description string
	// ParametersSchema is the JSON Schema of the tool's arguments.
	ParametersSchema map[string]any
}

// LLMRequest is one model invocation.
type LLMRequest struct {
	Model    string
	Messages []LLMMessage
	Tools    []ToolDefinition
	// ResponseSchema constrains the model to the structured response contract,
	// which is validated before anything reaches the frontend (§74, §178).
	ResponseSchema map[string]any
	MaxTokens      int
	Temperature    *float64
}

// ToolInvocation is a tool call the model requested.
type ToolInvocation struct {
	ID   string
	Name string
	// Arguments is the raw JSON argument object, validated by the orchestrator
	// before execution.
	Arguments []byte
}

// TokenUsage reports model consumption for cost observability (§122, §176).
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// LLMResponse is a completed model result.
type LLMResponse struct {
	// Content is the raw model output, still unvalidated at this layer.
	Content string
	// ToolCalls is non-empty when the model asked for tool execution instead
	// of producing a final answer.
	ToolCalls []ToolInvocation
	Usage     TokenUsage
	Model     string
}

// StreamSink receives incremental model output. The orchestrator adapts it
// into the SSE event contract; the provider knows nothing about HTTP (§82).
type StreamSink interface {
	// OnTextDelta receives an incremental piece of the answer.
	OnTextDelta(ctx context.Context, delta string) error
	// OnToolCall receives a completed tool-call request from the model.
	OnToolCall(ctx context.Context, invocation ToolInvocation) error
	// OnCompleted receives the final result once the stream ends.
	OnCompleted(ctx context.Context, response LLMResponse) error
}

// LLMProvider is the port every model vendor implements (§67). The concrete
// model is configuration and is never hardcoded in domain logic.
type LLMProvider interface {
	// Name identifies the implementation for logging and metrics.
	Name() Name
	// Generate performs a single non-streaming completion.
	Generate(ctx context.Context, req LLMRequest) (LLMResponse, error)
	// Stream performs a completion, emitting incremental output to sink.
	// Cancelling ctx must abort the upstream request (§82).
	Stream(ctx context.Context, req LLMRequest, sink StreamSink) error
}
