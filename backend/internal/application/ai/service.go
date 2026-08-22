// Package ai defines the AI application service, the orchestrator and the
// stream abstraction. The orchestrator routes intents and executes domain
// tools; it never performs financial calculations itself (§68).
package ai

import (
	"context"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/conversation"
	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// CreateConversationRequest starts a conversation about one wallet (§102).
type CreateConversationRequest struct {
	UserID   uuid.UUID
	WalletID uuid.UUID
	Title    *string
}

// SendMessageRequest is a user turn. Context narrows the question to a
// specific transaction or scenario when the UI initiated it from one.
type SendMessageRequest struct {
	UserID         uuid.UUID
	ConversationID uuid.UUID
	Content        string
	TransactionID  *uuid.UUID
	ScenarioID     *uuid.UUID
}

// Service owns conversations and message dispatch (§101, §103).
type Service interface {
	CreateConversation(ctx context.Context, req CreateConversationRequest) (conversation.Conversation, error)
	GetConversation(ctx context.Context, userID, conversationID uuid.UUID) (conversation.Conversation, error)
	// ListConversations returns the user's conversations, optionally narrowed
	// to one wallet.
	ListConversations(ctx context.Context, userID uuid.UUID, walletID *uuid.UUID, page shared.Cursor, limit int) (shared.Page[conversation.Conversation], error)
	// ListMessages returns conversation history newest first (§101).
	ListMessages(ctx context.Context, userID, conversationID uuid.UUID, page shared.Cursor, limit int) (shared.Page[conversation.Message], error)
	// SendMessage reserves an AI usage unit, persists the user turn and drives
	// the orchestrator, emitting events to sink as the answer is produced
	// (§80, §86, §103).
	SendMessage(ctx context.Context, req SendMessageRequest, sink StreamSink) error
}

// Orchestrator determines intent, selects and executes tools, builds the
// minimal context, invokes the LLM and validates the structured response
// (§68). It performs no financial calculations of its own (§50).
type Orchestrator interface {
	// Run answers one question. The tool loop is bounded by configuration so
	// the model cannot call tools indefinitely (§177).
	Run(ctx context.Context, req OrchestrationRequest, sink StreamSink) (RunResult, error)
}

// RunResult is the orchestrator output persisted on the assistant message.
type RunResult struct {
	Response  conversation.Response
	ToolCalls []conversation.ToolCall
	Content   string
}

// OrchestrationRequest is one AI turn together with everything the
// orchestrator is allowed to consider.
type OrchestrationRequest struct {
	UserID   uuid.UUID
	WalletID uuid.UUID
	Question string
	// History is the trailing window of prior turns. Unlimited history is
	// never sent to the model (§79).
	History []conversation.Message
	// TransactionID and ScenarioID scope the question when the UI supplied a
	// specific subject.
	TransactionID *uuid.UUID
	ScenarioID    *uuid.UUID
}

// Tool is one domain capability the model may invoke (§70). The backend
// executes it and returns structured facts; raw database rows, provider DTOs
// and unrelated user data never enter the result (§71).
type Tool interface {
	Name() conversation.ToolName
	Description() string
	// ParametersSchema is the JSON Schema of the tool's arguments.
	ParametersSchema() map[string]any
	// Execute runs the tool for the given user and returns a JSON document of
	// AI-facing facts.
	Execute(ctx context.Context, invocation ToolInvocation) ([]byte, error)
}

// ToolInvocation is a validated request to run one tool.
type ToolInvocation struct {
	ID        string
	UserID    uuid.UUID
	WalletID  uuid.UUID
	Arguments []byte
}

// ToolRegistry resolves tools by name. Intent routing uses a registry rather
// than a large hardcoded if/else tree (§69).
type ToolRegistry interface {
	Get(name conversation.ToolName) (Tool, bool)
	All() []Tool
}
