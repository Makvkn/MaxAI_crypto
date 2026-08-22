package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	appwallets "github.com/maxaicrypto/backend/internal/application/wallets"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/domain/wallet"
	"github.com/maxaicrypto/backend/internal/transport/http/middleware"
	"github.com/maxaicrypto/backend/internal/transport/http/request"
	"github.com/maxaicrypto/backend/internal/transport/http/response"
)

// WalletsHandler serves the /api/v1/wallets routes.
type WalletsHandler struct {
	service appwallets.Service
}

// NewWalletsHandler builds the wallets handler.
func NewWalletsHandler(service appwallets.Service) *WalletsHandler {
	return &WalletsHandler{service: service}
}

type createWalletRequest struct {
	ChainID chain.ID `json:"chain_id"`
	Address string   `json:"address"`
	Label   *string  `json:"label"`
}

type walletSyncErrorResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type walletSyncStateResponse struct {
	Status          wallet.SyncStatus       `json:"status"`
	Stage           *wallet.SyncStage       `json:"stage"`
	StagesCompleted []wallet.SyncStage      `json:"stages_completed"`
	StartedAt       *time.Time              `json:"started_at"`
	CompletedAt     *time.Time              `json:"completed_at"`
	LastSyncedAt    *time.Time              `json:"last_synced_at"`
	DataFreshness   *shared.DataFreshness   `json:"data_freshness"`
	Error           *walletSyncErrorResponse `json:"error"`
}

type walletResponse struct {
	ID        uuid.UUID               `json:"id"`
	ChainID   chain.ID                `json:"chain_id"`
	Address   string                  `json:"address"`
	Label     *string                 `json:"label"`
	Status    wallet.Status           `json:"status"`
	Sync      walletSyncStateResponse `json:"sync"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
}

// List handles GET /wallets.
func (h *WalletsHandler) List(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.service.List(r.Context(), userID, page.Cursor, page.Limit)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	items := make([]walletResponse, len(result.Items))
	for i, view := range result.Items {
		items[i] = mapWallet(view)
	}
	response.OK(w, r, shared.Page[walletResponse]{
		Items:      items,
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
	})
}

// Create handles POST /wallets.
func (h *WalletsHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		response.Error(w, r, apperr.New(apperr.CodeAuthentication))
		return
	}

	var body createWalletRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.Error(w, r, err)
		return
	}
	if body.ChainID == "" {
		response.Error(w, r, apperr.New(apperr.CodeValidation).
			WithMessage("The request is invalid.").
			WithDetail("fields", map[string]string{"chain_id": "is required"}))
		return
	}
	if body.Address == "" {
		response.Error(w, r, apperr.New(apperr.CodeValidation).
			WithMessage("The request is invalid.").
			WithDetail("fields", map[string]string{"address": "is required"}))
		return
	}

	view, err := h.service.Create(r.Context(), appwallets.CreateRequest{
		UserID:  userID,
		ChainID: body.ChainID,
		Address: body.Address,
		Label:   body.Label,
	})
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.Created(w, r, mapWallet(view))
}

// Get handles GET /wallets/{walletID}.
func (h *WalletsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		response.Error(w, r, apperr.New(apperr.CodeAuthentication))
		return
	}
	walletID, err := request.UUIDParam(r.PathValue("walletID"), "wallet_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	view, err := h.service.Get(r.Context(), userID, walletID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, r, mapWallet(view))
}

// Delete handles DELETE /wallets/{walletID}.
func (h *WalletsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		response.Error(w, r, apperr.New(apperr.CodeAuthentication))
		return
	}
	walletID, err := request.UUIDParam(r.PathValue("walletID"), "wallet_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}
	if err := h.service.Delete(r.Context(), userID, walletID); err != nil {
		response.Error(w, r, err)
		return
	}
	response.NoContent(w)
}

// RequestSync handles POST /wallets/{walletID}/sync.
func (h *WalletsHandler) RequestSync(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		response.Error(w, r, apperr.New(apperr.CodeAuthentication))
		return
	}
	walletID, err := request.UUIDParam(r.PathValue("walletID"), "wallet_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	view, err := h.service.RequestSync(r.Context(), userID, walletID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.JSON(w, r, http.StatusAccepted, mapWallet(view))
}

func mapWallet(view appwallets.View) walletResponse {
	sync := view.Sync
	stages := sync.StagesCompleted
	if stages == nil {
		stages = []wallet.SyncStage{}
	}
	resp := walletResponse{
		ID:        view.Wallet.ID,
		ChainID:   view.Wallet.ChainID,
		Address:   view.Wallet.Address,
		Label:     view.Wallet.Label,
		Status:    view.Wallet.Status,
		CreatedAt: view.Wallet.CreatedAt,
		UpdatedAt: view.Wallet.UpdatedAt,
		Sync: walletSyncStateResponse{
			Status:          sync.Status,
			Stage:           sync.Stage,
			StagesCompleted: stages,
			StartedAt:       sync.StartedAt,
			CompletedAt:     sync.CompletedAt,
			LastSyncedAt:    sync.LastSyncedAt,
			DataFreshness:   sync.DataFreshness,
		},
	}
	if sync.ErrorCode != nil {
		code := apperr.Code(*sync.ErrorCode)
		message := ""
		if sync.ErrorMessage != nil {
			message = *sync.ErrorMessage
		} else if apperr.IsKnown(code) {
			message = apperr.From(apperr.New(code)).Message
		}
		resp.Sync.Error = &walletSyncErrorResponse{
			Code:    string(code),
			Message: message,
			Details: map[string]any{},
		}
	}
	return resp
}
