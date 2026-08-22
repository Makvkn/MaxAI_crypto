// Package conversationrepo implements conversation persistence with sqlc.
package conversationrepo

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/maxaicrypto/backend/internal/domain/conversation"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/generated/sqlc"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
)

const previewMaxRunes = 120

// Repository implements conversation.Repository.
type Repository struct {
	pool *postgres.Pool
	tx   *postgres.TxRunner
}

// NewRepository builds a conversation repository.
func NewRepository(pool *postgres.Pool, tx *postgres.TxRunner) *Repository {
	return &Repository{pool: pool, tx: tx}
}

func (r *Repository) db(ctx context.Context) postgres.DBTX {
	if tx, ok := postgres.TxFrom(ctx); ok {
		return tx
	}
	return r.pool
}

func (r *Repository) queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(r.db(ctx))
}

// Create implements conversation.Repository.
func (r *Repository) Create(ctx context.Context, c conversation.Conversation) (conversation.Conversation, error) {
	row, err := r.queries(ctx).CreateConversation(ctx, sqlc.CreateConversationParams{
		UserID:   c.UserID,
		WalletID: c.WalletID,
		Title:    c.Title,
	})
	if err != nil {
		return conversation.Conversation{}, postgres.MapError(err)
	}
	return mapConversation(row), nil
}

// GetByID implements conversation.Repository.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (conversation.Conversation, error) {
	row, err := r.queries(ctx).GetConversationByID(ctx, id)
	if err != nil {
		return conversation.Conversation{}, postgres.MapError(err)
	}
	return mapConversation(row), nil
}

// ListByUser implements conversation.Repository.
func (r *Repository) ListByUser(ctx context.Context, userID uuid.UUID, walletID *uuid.UUID, page shared.Cursor, limit int) ([]conversation.Conversation, error) {
	params := sqlc.ListConversationsByUserParams{
		UserID: userID,
		Limit:  int32(limit),
	}
	if walletID != nil {
		params.WalletID = uuid.NullUUID{UUID: *walletID, Valid: true}
	}
	if !page.IsZero() {
		cursorID, err := uuid.Parse(page.TieBreaker)
		if err != nil {
			return nil, err
		}
		params.CursorAt = &page.SortKey
		params.CursorID = uuid.NullUUID{UUID: cursorID, Valid: true}
	}

	rows, err := r.queries(ctx).ListConversationsByUser(ctx, params)
	if err != nil {
		return nil, postgres.MapError(err)
	}
	conversations := make([]conversation.Conversation, len(rows))
	for i, row := range rows {
		conversations[i] = mapConversation(row)
	}
	return conversations, nil
}

// AppendMessage implements conversation.Repository.
func (r *Repository) AppendMessage(ctx context.Context, m conversation.Message) (conversation.Message, error) {
	var created conversation.Message
	err := r.tx.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := sqlc.New(tx)
		row, err := insertMessage(ctx, q, m)
		if err != nil {
			return err
		}
		created = row

		preview := previewText(m)
		if err := q.BumpConversationOnMessage(ctx, sqlc.BumpConversationOnMessageParams{
			ID:                 m.ConversationID,
			LastMessagePreview: preview,
		}); err != nil {
			return postgres.MapError(err)
		}
		return nil
	})
	return created, err
}

// UpdateMessage implements conversation.Repository.
func (r *Repository) UpdateMessage(ctx context.Context, m conversation.Message) error {
	responseJSON, toolCallsJSON, err := encodeMessagePayload(m)
	if err != nil {
		return err
	}
	return postgres.MapError(r.queries(ctx).UpdateConversationMessage(ctx, sqlc.UpdateConversationMessageParams{
		ID:           m.ID,
		Status:       string(m.Status),
		Content:      m.Content,
		Response:     responseJSON,
		ToolCalls:    toolCallsJSON,
		ErrorCode:    m.ErrorCode,
		ErrorMessage: m.ErrorMessage,
	}))
}

// ListMessages implements conversation.Repository.
func (r *Repository) ListMessages(ctx context.Context, conversationID uuid.UUID, page shared.Cursor, limit int) ([]conversation.Message, error) {
	params := sqlc.ListConversationMessagesParams{
		ConversationID: conversationID,
		Limit:          int32(limit),
	}
	if !page.IsZero() {
		cursorID, err := uuid.Parse(page.TieBreaker)
		if err != nil {
			return nil, err
		}
		params.CursorAt = &page.SortKey
		params.CursorID = uuid.NullUUID{UUID: cursorID, Valid: true}
	}

	rows, err := r.queries(ctx).ListConversationMessages(ctx, params)
	if err != nil {
		return nil, postgres.MapError(err)
	}
	return mapMessages(rows)
}

// ListRecentMessages implements conversation.Repository.
func (r *Repository) ListRecentMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]conversation.Message, error) {
	rows, err := r.queries(ctx).ListRecentConversationMessages(ctx, sqlc.ListRecentConversationMessagesParams{
		ConversationID: conversationID,
		Limit:          int32(limit),
	})
	if err != nil {
		return nil, postgres.MapError(err)
	}
	messages, err := mapMessages(rows)
	if err != nil {
		return nil, err
	}
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

func insertMessage(ctx context.Context, q *sqlc.Queries, m conversation.Message) (conversation.Message, error) {
	responseJSON, toolCallsJSON, err := encodeMessagePayload(m)
	if err != nil {
		return conversation.Message{}, err
	}
	row, err := q.InsertConversationMessage(ctx, sqlc.InsertConversationMessageParams{
		ConversationID: m.ConversationID,
		Role:           string(m.Role),
		Status:         string(m.Status),
		Content:        m.Content,
		Response:       responseJSON,
		ToolCalls:      toolCallsJSON,
		ErrorCode:      m.ErrorCode,
		ErrorMessage:   m.ErrorMessage,
	})
	if err != nil {
		return conversation.Message{}, postgres.MapError(err)
	}
	return mapMessage(row)
}

func encodeMessagePayload(m conversation.Message) ([]byte, []byte, error) {
	var responseJSON []byte
	if m.Response != nil {
		payload, err := json.Marshal(m.Response)
		if err != nil {
			return nil, nil, err
		}
		responseJSON = payload
	}
	toolCalls := m.ToolCalls
	if toolCalls == nil {
		toolCalls = []conversation.ToolCall{}
	}
	toolCallsJSON, err := json.Marshal(toolCalls)
	if err != nil {
		return nil, nil, err
	}
	return responseJSON, toolCallsJSON, nil
}

func mapConversation(row sqlc.Conversation) conversation.Conversation {
	return conversation.Conversation{
		ID:                 row.ID,
		UserID:             row.UserID,
		WalletID:           row.WalletID,
		Title:              row.Title,
		MessageCount:       int(row.MessageCount),
		LastMessagePreview: row.LastMessagePreview,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}

func mapMessages(rows []sqlc.ConversationMessage) ([]conversation.Message, error) {
	messages := make([]conversation.Message, len(rows))
	for i, row := range rows {
		message, err := mapMessage(row)
		if err != nil {
			return nil, err
		}
		messages[i] = message
	}
	return messages, nil
}

func mapMessage(row sqlc.ConversationMessage) (conversation.Message, error) {
	message := conversation.Message{
		ID:             row.ID,
		ConversationID: row.ConversationID,
		Role:           conversation.Role(row.Role),
		Status:         conversation.MessageStatus(row.Status),
		Content:        row.Content,
		ErrorCode:      row.ErrorCode,
		ErrorMessage:   row.ErrorMessage,
		CreatedAt:      row.CreatedAt,
	}
	if len(row.Response) > 0 {
		var response conversation.Response
		if err := json.Unmarshal(row.Response, &response); err != nil {
			return conversation.Message{}, err
		}
		message.Response = &response
	}
	if len(row.ToolCalls) > 0 {
		if err := json.Unmarshal(row.ToolCalls, &message.ToolCalls); err != nil {
			return conversation.Message{}, err
		}
	}
	if message.ToolCalls == nil {
		message.ToolCalls = []conversation.ToolCall{}
	}
	return message, nil
}

func previewText(m conversation.Message) *string {
	content := m.Content
	if m.Role == conversation.RoleAssistant && m.Response != nil && m.Response.Answer != "" {
		content = m.Response.Answer
	}
	if content == "" {
		return nil
	}
	runes := []rune(content)
	if len(runes) > previewMaxRunes {
		truncated := string(runes[:previewMaxRunes]) + "…"
		return &truncated
	}
	return &content
}

var _ conversation.Repository = (*Repository)(nil)
