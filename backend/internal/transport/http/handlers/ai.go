package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	appai "github.com/maxaicrypto/backend/internal/application/ai"
	appscenarios "github.com/maxaicrypto/backend/internal/application/scenarios"
	appusage "github.com/maxaicrypto/backend/internal/application/usage"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/conversation"
	"github.com/maxaicrypto/backend/internal/domain/subscription"
	"github.com/maxaicrypto/backend/internal/domain/usage"
	"github.com/maxaicrypto/backend/internal/transport/http/middleware"
	"github.com/maxaicrypto/backend/internal/transport/http/request"
	"github.com/maxaicrypto/backend/internal/transport/http/response"
)

// AIHandler serves AI routes.
type AIHandler struct {
	ai        appai.Service
	usage     appusage.Service
	scenarios appscenarios.Service
}

// NewAIHandler builds the AI handler.
func NewAIHandler(ai appai.Service, usage appusage.Service, scenarios appscenarios.Service) *AIHandler {
	return &AIHandler{ai: ai, usage: usage, scenarios: scenarios}
}

type aiUsageResponse struct {
	Date      string             `json:"date"`
	Used      int                `json:"used"`
	Limit     int                `json:"limit"`
	Remaining int                `json:"remaining"`
	ResetsAt  time.Time          `json:"resets_at"`
	Plan      subscription.Plan  `json:"plan"`
}

type conversationResponse struct {
	ID                 uuid.UUID `json:"id"`
	WalletID           uuid.UUID `json:"wallet_id"`
	Title              string    `json:"title"`
	MessageCount       int       `json:"message_count"`
	LastMessagePreview *string   `json:"last_message_preview"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type aiClaimResponse struct {
	Text     string              `json:"text"`
	Evidence []aiEvidenceResponse `json:"evidence"`
}

type aiEvidenceResponse struct {
	Type conversation.EvidenceType `json:"type"`
	ID   string                    `json:"id"`
}

type aiReferenceResponse struct {
	Type  conversation.ReferenceType `json:"type"`
	ID    string                     `json:"id"`
	Label *string                    `json:"label"`
}

type aiResponseBody struct {
	Answer            string                `json:"answer"`
	Intent            conversation.Intent   `json:"intent"`
	DataQuality       string                `json:"data_quality"`
	Claims            []aiClaimResponse     `json:"claims"`
	References        []aiReferenceResponse `json:"references"`
	UnsupportedReason *string               `json:"unsupported_reason"`
}

type aiToolCallResponse struct {
	ID          string                      `json:"id"`
	Tool        conversation.ToolName       `json:"tool"`
	Status      conversation.ToolCallStatus `json:"status"`
	StartedAt   time.Time                   `json:"started_at"`
	CompletedAt *time.Time                  `json:"completed_at"`
}

type apiErrorBodyResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
}

type conversationMessageResponse struct {
	ID             uuid.UUID              `json:"id"`
	ConversationID uuid.UUID              `json:"conversation_id"`
	Role           conversation.Role      `json:"role"`
	Status         conversation.MessageStatus `json:"status"`
	Content        string                 `json:"content"`
	Response       *aiResponseBody        `json:"response"`
	ToolCalls      []aiToolCallResponse   `json:"tool_calls"`
	Error          *apiErrorBodyResponse  `json:"error"`
	CreatedAt      time.Time              `json:"created_at"`
}

type createConversationRequest struct {
	WalletID uuid.UUID `json:"wallet_id"`
	Title    *string   `json:"title"`
}

type sendMessageRequest struct {
	Content string `json:"content"`
	Context *struct {
		TransactionID *uuid.UUID `json:"transaction_id"`
		ScenarioID    *uuid.UUID `json:"scenario_id"`
	} `json:"context"`
}

// GetUsage handles GET /ai/usage.
func (h *AIHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		response.Error(w, r, apperr.New(apperr.CodeAuthentication))
		return
	}
	daily, err := h.usage.Today(r.Context(), userID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, r, mapAIUsage(daily))
}

// ListConversations handles GET /ai/conversations.
func (h *AIHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		response.Error(w, r, apperr.New(apperr.CodeAuthentication))
		return
	}
	page, err := request.ParsePage(r)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	var walletID *uuid.UUID
	if raw, ok := request.QueryString(r, "wallet_id"); ok {
		id, err := uuid.Parse(raw)
		if err != nil {
			response.Error(w, r, apperr.New(apperr.CodeValidation).
				WithMessage("The wallet_id parameter is not valid.").
				WithDetail("fields", map[string]string{"wallet_id": "must be a valid identifier"}))
			return
		}
		walletID = &id
	}

	result, err := h.ai.ListConversations(r.Context(), userID, walletID, page.Cursor, page.Limit)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	items := make([]conversationResponse, len(result.Items))
	for i, item := range result.Items {
		items[i] = mapConversation(item)
	}
	response.OK(w, r, Page[conversationResponse]{
		Items:      items,
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
	})
}

// CreateConversation handles POST /ai/conversations.
func (h *AIHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		response.Error(w, r, apperr.New(apperr.CodeAuthentication))
		return
	}
	var body createConversationRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.Error(w, r, err)
		return
	}
	if body.WalletID == uuid.Nil {
		response.Error(w, r, apperr.New(apperr.CodeValidation).
			WithMessage("The wallet_id field is required.").
			WithDetail("fields", map[string]string{"wallet_id": "is required"}))
		return
	}

	conv, err := h.ai.CreateConversation(r.Context(), appai.CreateConversationRequest{
		UserID:   userID,
		WalletID: body.WalletID,
		Title:    body.Title,
	})
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, r, mapConversation(conv))
}

// GetConversation handles GET /ai/conversations/{conversationID}.
func (h *AIHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		response.Error(w, r, apperr.New(apperr.CodeAuthentication))
		return
	}
	conversationID, err := request.UUIDParam(r.PathValue("conversationID"), "conversation_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	conv, err := h.ai.GetConversation(r.Context(), userID, conversationID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, r, mapConversation(conv))
}

// ListMessages handles GET /ai/conversations/{conversationID}/messages.
func (h *AIHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		response.Error(w, r, apperr.New(apperr.CodeAuthentication))
		return
	}
	conversationID, err := request.UUIDParam(r.PathValue("conversationID"), "conversation_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	page, err := request.ParsePage(r)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	result, err := h.ai.ListMessages(r.Context(), userID, conversationID, page.Cursor, page.Limit)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	items := make([]conversationMessageResponse, len(result.Items))
	for i, item := range result.Items {
		items[i] = mapConversationMessage(item)
	}
	response.OK(w, r, Page[conversationMessageResponse]{
		Items:      items,
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
	})
}

// SendMessage handles POST /ai/conversations/{conversationID}/messages.
func (h *AIHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		response.Error(w, r, apperr.New(apperr.CodeAuthentication))
		return
	}
	conversationID, err := request.UUIDParam(r.PathValue("conversationID"), "conversation_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	var body sendMessageRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.Error(w, r, err)
		return
	}

	req := appai.SendMessageRequest{
		UserID:         userID,
		ConversationID: conversationID,
		Content:        body.Content,
	}
	if body.Context != nil {
		req.TransactionID = body.Context.TransactionID
		req.ScenarioID = body.Context.ScenarioID
	}

	response.BeginStream(w)
	sink := &httpStreamSink{writer: w}
	if err := h.ai.SendMessage(r.Context(), req, sink); err != nil {
		if !sink.terminal {
			appErr := apperr.From(err)
			if appErr == nil {
				appErr = apperr.New(apperr.CodeInternal)
			}
			details := appErr.Details
			if details == nil {
				details = map[string]any{}
			}
			_ = response.Stream(w, string(appai.EventError), map[string]any{
				"type": "error",
				"error": apiErrorBodyResponse{
					Code:    string(appErr.Code),
					Message: appErr.Message,
					Details: details,
				},
			})
		}
	}
}

type httpStreamSink struct {
	writer   http.ResponseWriter
	terminal bool
}

func (s *httpStreamSink) Send(ctx context.Context, event appai.Event) error {
	_ = ctx
	switch event.Type {
	case appai.EventCompleted:
		s.terminal = true
		payload := map[string]any{
			"type": "completed",
		}
		if event.Message != nil {
			payload["message"] = mapConversationMessage(*event.Message)
		}
		if event.Usage != nil {
			payload["usage"] = mapAIUsage(*event.Usage)
		} else {
			payload["usage"] = nil
		}
		return response.Stream(s.writer, string(event.Type), payload)
	case appai.EventError:
		s.terminal = true
		return response.Stream(s.writer, string(event.Type), map[string]any{
			"type": "error",
			"error": apiErrorBodyResponse{
				Code:    event.ErrorCode,
				Message: event.ErrorMessage,
				Details: map[string]any{},
			},
		})
	default:
		payload := map[string]any{"type": string(event.Type)}
		switch event.Type {
		case appai.EventToolStarted:
			payload["tool_call_id"] = event.ToolCallID
			payload["tool"] = event.Tool
		case appai.EventToolCompleted:
			payload["tool_call_id"] = event.ToolCallID
			payload["tool"] = event.Tool
			payload["ok"] = event.ToolOK
		case appai.EventTextDelta:
			payload["text"] = event.Text
		}
		return response.Stream(s.writer, string(event.Type), payload)
	}
}

func mapAIUsage(daily usage.Daily) aiUsageResponse {
	return aiUsageResponse{
		Date:      daily.Date.Format("2006-01-02"),
		Used:      daily.Used,
		Limit:     daily.Limit,
		Remaining: daily.Remaining(),
		ResetsAt:  daily.ResetsAt,
		Plan:      daily.Plan,
	}
}

func mapConversation(c conversation.Conversation) conversationResponse {
	return conversationResponse{
		ID:                 c.ID,
		WalletID:           c.WalletID,
		Title:              c.Title,
		MessageCount:       c.MessageCount,
		LastMessagePreview: c.LastMessagePreview,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
	}
}

func mapConversationMessage(m conversation.Message) conversationMessageResponse {
	toolCalls := make([]aiToolCallResponse, len(m.ToolCalls))
	for i, call := range m.ToolCalls {
		toolCalls[i] = aiToolCallResponse{
			ID:          call.ID,
			Tool:        call.Tool,
			Status:      call.Status,
			StartedAt:   call.StartedAt,
			CompletedAt: call.CompletedAt,
		}
	}
	if toolCalls == nil {
		toolCalls = []aiToolCallResponse{}
	}

	var messageError *apiErrorBodyResponse
	if m.ErrorCode != nil {
		msg := ""
		if m.ErrorMessage != nil {
			msg = *m.ErrorMessage
		}
		messageError = &apiErrorBodyResponse{
			Code:    *m.ErrorCode,
			Message: msg,
			Details: map[string]any{},
		}
	}

	return conversationMessageResponse{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		Role:           m.Role,
		Status:         m.Status,
		Content:        m.Content,
		Response:       mapAIResponse(m.Response),
		ToolCalls:      toolCalls,
		Error:          messageError,
		CreatedAt:      m.CreatedAt,
	}
}

func mapAIResponse(resp *conversation.Response) *aiResponseBody {
	if resp == nil {
		return nil
	}
	claims := make([]aiClaimResponse, len(resp.Claims))
	for i, claim := range resp.Claims {
		evidence := make([]aiEvidenceResponse, len(claim.Evidence))
		for j, item := range claim.Evidence {
			evidence[j] = aiEvidenceResponse{Type: item.Type, ID: item.ID}
		}
		claims[i] = aiClaimResponse{Text: claim.Text, Evidence: evidence}
	}
	if claims == nil {
		claims = []aiClaimResponse{}
	}
	references := make([]aiReferenceResponse, len(resp.References))
	for i, ref := range resp.References {
		references[i] = aiReferenceResponse{Type: ref.Type, ID: ref.ID, Label: ref.Label}
	}
	if references == nil {
		references = []aiReferenceResponse{}
	}
	return &aiResponseBody{
		Answer:            resp.Answer,
		Intent:            resp.Intent,
		DataQuality:       string(resp.DataQuality),
		Claims:            claims,
		References:        references,
		UnsupportedReason: resp.UnsupportedReason,
	}
}

// Page mirrors shared.Page with concrete item type for JSON encoding.
type Page[T any] struct {
	Items      []T     `json:"items"`
	NextCursor *string `json:"next_cursor"`
	HasMore    bool    `json:"has_more"`
}
