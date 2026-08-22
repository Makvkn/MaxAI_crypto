package conversation

import (
	"context"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// Repository persists conversations and messages.
type Repository interface {
	Create(ctx context.Context, c Conversation) (Conversation, error)
	// GetByID returns a conversation. Ownership is enforced by the application
	// layer against the requesting user (§181).
	GetByID(ctx context.Context, id uuid.UUID) (Conversation, error)
	// ListByUser returns conversations most recently updated first, optionally
	// narrowed to one wallet.
	ListByUser(ctx context.Context, userID uuid.UUID, walletID *uuid.UUID, page shared.Cursor, limit int) ([]Conversation, error)

	// AppendMessage persists a message and updates the conversation's
	// denormalized preview counters in one transaction.
	AppendMessage(ctx context.Context, m Message) (Message, error)
	// UpdateMessage persists the outcome of a streamed assistant message.
	UpdateMessage(ctx context.Context, m Message) error
	// ListMessages returns messages newest first for cursor pagination (§101).
	ListMessages(ctx context.Context, conversationID uuid.UUID, page shared.Cursor, limit int) ([]Message, error)
	// ListRecentMessages returns the trailing window of messages used to build
	// LLM context. Unlimited history is never sent to the model (§79).
	ListRecentMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]Message, error)
}
