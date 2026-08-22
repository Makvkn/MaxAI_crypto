package ai

import (
	"context"
	"strings"

	"github.com/google/uuid"

	appusage "github.com/maxaicrypto/backend/internal/application/usage"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/conversation"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/domain/usage"
	"github.com/maxaicrypto/backend/internal/domain/wallet"
)

const (
	defaultConversationTitle = "Portfolio analysis"
	maxContextMessages       = 12
)

// App implements Service.
type App struct {
	conversations conversation.Repository
	wallets       wallet.Repository
	syncStates    wallet.SyncStateRepository
	usage         appusage.Service
	orchestrator  Orchestrator
}

// NewApp wires the AI application service.
func NewApp(
	conversations conversation.Repository,
	wallets wallet.Repository,
	syncStates wallet.SyncStateRepository,
	usage appusage.Service,
	orchestrator Orchestrator,
) *App {
	return &App{
		conversations: conversations,
		wallets:       wallets,
		syncStates:    syncStates,
		usage:         usage,
		orchestrator:  orchestrator,
	}
}

// CreateConversation implements Service.
func (a *App) CreateConversation(ctx context.Context, req CreateConversationRequest) (conversation.Conversation, error) {
	if err := a.requireReadyWallet(ctx, req.UserID, req.WalletID); err != nil {
		return conversation.Conversation{}, err
	}

	title := defaultConversationTitle
	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		if trimmed != "" {
			title = trimmed
		}
	}

	return a.conversations.Create(ctx, conversation.Conversation{
		UserID:   req.UserID,
		WalletID: req.WalletID,
		Title:    title,
	})
}

// GetConversation implements Service.
func (a *App) GetConversation(ctx context.Context, userID, conversationID uuid.UUID) (conversation.Conversation, error) {
	conv, err := a.conversations.GetByID(ctx, conversationID)
	if err != nil {
		if appErr := apperr.From(err); appErr != nil && appErr.Code == apperr.CodeNotFound {
			return conversation.Conversation{}, apperr.New(apperr.CodeConversationNotFound)
		}
		return conversation.Conversation{}, err
	}
	if conv.UserID != userID {
		return conversation.Conversation{}, apperr.New(apperr.CodeConversationNotFound)
	}
	return conv, nil
}

// ListConversations implements Service.
func (a *App) ListConversations(ctx context.Context, userID uuid.UUID, walletID *uuid.UUID, page shared.Cursor, limit int) (shared.Page[conversation.Conversation], error) {
	if limit < 1 {
		limit = 1
	}
	rows, err := a.conversations.ListByUser(ctx, userID, walletID, page, limit+1)
	if err != nil {
		return shared.Page[conversation.Conversation]{}, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	var next shared.Cursor
	if hasMore {
		last := rows[len(rows)-1]
		next = shared.NewCursor(last.UpdatedAt, last.ID.String())
	}
	return shared.NewPage(rows, next, hasMore), nil
}

// ListMessages implements Service.
func (a *App) ListMessages(ctx context.Context, userID, conversationID uuid.UUID, page shared.Cursor, limit int) (shared.Page[conversation.Message], error) {
	if _, err := a.GetConversation(ctx, userID, conversationID); err != nil {
		return shared.Page[conversation.Message]{}, err
	}
	if limit < 1 {
		limit = 1
	}
	rows, err := a.conversations.ListMessages(ctx, conversationID, page, limit+1)
	if err != nil {
		return shared.Page[conversation.Message]{}, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	var next shared.Cursor
	if hasMore {
		last := rows[len(rows)-1]
		next = shared.NewCursor(last.CreatedAt, last.ID.String())
	}
	return shared.NewPage(rows, next, hasMore), nil
}

// SendMessage implements Service.
func (a *App) SendMessage(ctx context.Context, req SendMessageRequest, sink StreamSink) error {
	conv, err := a.GetConversation(ctx, req.UserID, req.ConversationID)
	if err != nil {
		return err
	}
	if err := a.requireReadyWallet(ctx, req.UserID, conv.WalletID); err != nil {
		return err
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return apperr.New(apperr.CodeValidation).
			WithMessage("Message content is required.").
			WithDetail("fields", map[string]string{"content": "is required"})
	}

	reservation, err := a.usage.Reserve(ctx, req.UserID, usage.OperationAsk, uuid.New().String())
	if err != nil {
		return err
	}

	userMessage, err := a.conversations.AppendMessage(ctx, conversation.Message{
		ConversationID: req.ConversationID,
		Role:           conversation.RoleUser,
		Status:         conversation.MessageCompleted,
		Content:        content,
	})
	if err != nil {
		_ = a.usage.Release(ctx, reservation)
		return err
	}
	_ = userMessage

	assistant, err := a.conversations.AppendMessage(ctx, conversation.Message{
		ConversationID: req.ConversationID,
		Role:           conversation.RoleAssistant,
		Status:         conversation.MessageStreaming,
		Content:        "",
		ToolCalls:      []conversation.ToolCall{},
	})
	if err != nil {
		_ = a.usage.Release(ctx, reservation)
		return err
	}

	history, err := a.conversations.ListRecentMessages(ctx, req.ConversationID, maxContextMessages)
	if err != nil {
		_ = a.usage.Release(ctx, reservation)
		return err
	}

	collecting := &textCollector{inner: sink}
	result, err := a.orchestrator.Run(ctx, OrchestrationRequest{
		UserID:        req.UserID,
		WalletID:      conv.WalletID,
		Question:      content,
		History:       history,
		TransactionID: req.TransactionID,
		ScenarioID:    req.ScenarioID,
	}, collecting)
	if err != nil {
		_ = a.failAssistant(ctx, assistant, err)
		_ = a.usage.Release(ctx, reservation)
		return a.emitTerminalError(ctx, sink, err)
	}

	assistant.Status = conversation.MessageCompleted
	assistant.Content = collecting.text.String()
	if assistant.Content == "" {
		assistant.Content = result.Content
	}
	assistant.Response = &result.Response
	assistant.ToolCalls = result.ToolCalls
	if err := a.conversations.UpdateMessage(ctx, assistant); err != nil {
		_ = a.usage.Release(ctx, reservation)
		return err
	}
	if err := a.usage.Commit(ctx, reservation); err != nil {
		return err
	}

	daily, err := a.usage.Today(ctx, req.UserID)
	if err != nil {
		return err
	}

	return sink.Send(ctx, Event{
		Type:    EventCompleted,
		Message: &assistant,
		Usage:   &daily,
	})
}

func (a *App) failAssistant(ctx context.Context, message conversation.Message, err error) error {
	code := string(apperr.CodeInternal)
	msg := "The AI request failed."
	if appErr := apperr.From(err); appErr != nil {
		code = string(appErr.Code)
		if appErr.Message != "" {
			msg = appErr.Message
		}
	}
	message.Status = conversation.MessageFailed
	message.ErrorCode = &code
	message.ErrorMessage = &msg
	return a.conversations.UpdateMessage(ctx, message)
}

func (a *App) emitTerminalError(ctx context.Context, sink StreamSink, err error) error {
	code := apperr.CodeInternal
	message := "The AI request failed."
	if appErr := apperr.From(err); appErr != nil {
		code = appErr.Code
		if appErr.Message != "" {
			message = appErr.Message
		}
	}
	return sink.Send(ctx, Event{
		Type:         EventError,
		ErrorCode:    string(code),
		ErrorMessage: message,
	})
}

func (a *App) requireReadyWallet(ctx context.Context, userID, walletID uuid.UUID) error {
	w, err := a.wallets.GetByID(ctx, walletID)
	if err != nil {
		if appErr := apperr.From(err); appErr != nil && appErr.Code == apperr.CodeNotFound {
			return apperr.New(apperr.CodeWalletNotFound)
		}
		return err
	}
	if w.UserID != userID {
		return apperr.New(apperr.CodeWalletNotFound)
	}

	syncState, err := a.syncStates.Get(ctx, walletID)
	if err != nil {
		return err
	}
	switch syncState.Status {
	case wallet.SyncPending, wallet.SyncSyncing:
		return apperr.New(apperr.CodeWalletNotReady).
			WithDetail("sync_status", string(syncState.Status))
	case wallet.SyncFailed:
		return apperr.New(apperr.CodeWalletSyncFailed)
	}
	return nil
}

type textCollector struct {
	inner StreamSink
	text  strings.Builder
}

func (c *textCollector) Send(ctx context.Context, event Event) error {
	if event.Type == EventTextDelta {
		c.text.WriteString(event.Text)
	}
	return c.inner.Send(ctx, event)
}

var _ Service = (*App)(nil)
