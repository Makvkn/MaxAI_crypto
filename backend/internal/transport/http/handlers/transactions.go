package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	apptransactions "github.com/maxaicrypto/backend/internal/application/transactions"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/domain/transaction"
	"github.com/maxaicrypto/backend/internal/transport/http/middleware"
	"github.com/maxaicrypto/backend/internal/transport/http/request"
	"github.com/maxaicrypto/backend/internal/transport/http/response"
)

// TransactionsHandler serves wallet transaction routes.
type TransactionsHandler struct {
	service apptransactions.Service
}

// NewTransactionsHandler builds the transactions handler.
func NewTransactionsHandler(service apptransactions.Service) *TransactionsHandler {
	return &TransactionsHandler{service: service}
}

type transactionResponse struct {
	ID           uuid.UUID              `json:"id"`
	WalletID     uuid.UUID              `json:"wallet_id"`
	ChainID      chain.ID               `json:"chain_id"`
	TxHash       string                 `json:"tx_hash"`
	BlockNumber  *int64                 `json:"block_number"`
	Timestamp    time.Time              `json:"timestamp"`
	Status       transaction.Status     `json:"status"`
	Type         transaction.Type       `json:"type"`
	FromAddress  *string                `json:"from_address"`
	ToAddress    *string                `json:"to_address"`
	AssetIn      *assetResponse         `json:"asset_in"`
	AmountIn     shared.NullDecimal     `json:"amount_in"`
	ValueInUSD   shared.NullDecimal     `json:"value_in_usd"`
	AssetOut     *assetResponse         `json:"asset_out"`
	AmountOut    shared.NullDecimal     `json:"amount_out"`
	ValueOutUSD  shared.NullDecimal     `json:"value_out_usd"`
	FeeAsset     *assetResponse         `json:"fee_asset"`
	FeeAmount    shared.NullDecimal     `json:"fee_amount"`
	FeeValueUSD  shared.NullDecimal     `json:"fee_value_usd"`
	Protocol     *string                `json:"protocol"`
	Counterparty *string                `json:"counterparty"`
	ExplorerURL  *string                `json:"explorer_url"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// List handles GET /wallets/{walletID}/transactions.
func (h *TransactionsHandler) List(w http.ResponseWriter, r *http.Request) {
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
	page, err := request.ParsePage(r)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	filter := transaction.Filter{}
	if raw, ok := request.QueryString(r, "type"); ok {
		txType := transaction.Type(raw)
		if !isKnownTransactionType(txType) {
			response.Error(w, r, apperr.New(apperr.CodeValidation).
				WithMessage("The type parameter is not valid.").
				WithDetail("fields", map[string]string{"type": "is not a valid transaction type"}))
			return
		}
		filter.Type = &txType
	}

	result, err := h.service.List(r.Context(), userID, walletID, filter, page.Cursor, page.Limit)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	items := make([]transactionResponse, len(result.Items))
	for i, view := range result.Items {
		items[i] = mapTransaction(view)
	}
	response.OK(w, r, shared.Page[transactionResponse]{
		Items:      items,
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
	})
}

// Get handles GET /wallets/{walletID}/transactions/{transactionID}.
func (h *TransactionsHandler) Get(w http.ResponseWriter, r *http.Request) {
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
	transactionID, err := request.UUIDParam(r.PathValue("transactionID"), "transaction_id")
	if err != nil {
		response.Error(w, r, err)
		return
	}

	view, err := h.service.Get(r.Context(), userID, walletID, transactionID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, r, mapTransaction(view))
}

func mapTransaction(view apptransactions.View) transactionResponse {
	tx := view.Transaction
	return transactionResponse{
		ID:           tx.ID,
		WalletID:     tx.WalletID,
		ChainID:      tx.ChainID,
		TxHash:       tx.TxHash,
		BlockNumber:  tx.BlockNumber,
		Timestamp:    tx.Timestamp,
		Status:       tx.Status,
		Type:         tx.Type,
		FromAddress:  tx.FromAddress,
		ToAddress:    tx.ToAddress,
		AssetIn:      mapAsset(view.AssetIn),
		AmountIn:     tx.AmountIn,
		ValueInUSD:   view.ValueInUSD,
		AssetOut:     mapAsset(view.AssetOut),
		AmountOut:    tx.AmountOut,
		ValueOutUSD:  view.ValueOutUSD,
		FeeAsset:     mapAsset(view.FeeAsset),
		FeeAmount:    tx.FeeAmount,
		FeeValueUSD:  view.FeeValueUSD,
		Protocol:     tx.Protocol,
		Counterparty: tx.Counterparty,
		ExplorerURL:  view.ExplorerURL,
		CreatedAt:    tx.CreatedAt,
		UpdatedAt:    tx.UpdatedAt,
	}
}

func mapAsset(ast *asset.Asset) *assetResponse {
	if ast == nil {
		return nil
	}
	return &assetResponse{
		ID:              ast.ID,
		ChainID:         ast.ChainID,
		ContractAddress: ast.ContractAddress,
		Symbol:          ast.Symbol,
		Name:            ast.Name,
		Decimals:        ast.Decimals,
		AssetType:       ast.Type,
		IconURL:         ast.IconURL,
		HasMarketData:   ast.HasMarketData(),
	}
}

func isKnownTransactionType(t transaction.Type) bool {
	switch t {
	case transaction.TypeTransfer,
		transaction.TypeSwap,
		transaction.TypeStake,
		transaction.TypeUnstake,
		transaction.TypeClaim,
		transaction.TypeApprove,
		transaction.TypeContractInteraction,
		transaction.TypeUnknown:
		return true
	default:
		return false
	}
}
