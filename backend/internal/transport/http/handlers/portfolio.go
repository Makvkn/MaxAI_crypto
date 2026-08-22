package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	appportfolio "github.com/maxaicrypto/backend/internal/application/portfolio"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/portfolio"
	"github.com/maxaicrypto/backend/internal/domain/price"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/transport/http/middleware"
	"github.com/maxaicrypto/backend/internal/transport/http/request"
	"github.com/maxaicrypto/backend/internal/transport/http/response"
)

// PortfolioHandler serves wallet portfolio routes.
type PortfolioHandler struct {
	service   appportfolio.Service
	freshness shared.FreshnessThresholds
}

// NewPortfolioHandler builds the portfolio handler.
func NewPortfolioHandler(service appportfolio.Service, freshness shared.FreshnessThresholds) *PortfolioHandler {
	return &PortfolioHandler{service: service, freshness: freshness}
}

type assetResponse struct {
	ID              uuid.UUID   `json:"id"`
	ChainID         chain.ID    `json:"chain_id"`
	ContractAddress *string     `json:"contract_address"`
	Symbol          string      `json:"symbol"`
	Name            string      `json:"name"`
	Decimals        int         `json:"decimals"`
	AssetType       asset.Type  `json:"asset_type"`
	IconURL         *string     `json:"icon_url"`
	HasMarketData   bool        `json:"has_market_data"`
}

type priceResponse struct {
	AssetID      uuid.UUID            `json:"asset_id"`
	ValueUSD     shared.NullDecimal   `json:"value_usd"`
	Currency     shared.Currency      `json:"currency"`
	Status       price.Status         `json:"status"`
	Freshness    shared.DataFreshness `json:"freshness"`
	AsOf         *time.Time           `json:"as_of"`
	Change24hPct shared.NullDecimal   `json:"change_24h_pct"`
}

type walletPositionResponse struct {
	Asset           assetResponse        `json:"asset"`
	Balance         shared.Decimal       `json:"balance"`
	BalanceRaw      string               `json:"balance_raw"`
	Price           *priceResponse       `json:"price"`
	ValueUSD        shared.NullDecimal   `json:"value_usd"`
	AllocationPct   shared.NullDecimal   `json:"allocation_pct"`
	Change24hPct    shared.NullDecimal   `json:"change_24h_pct"`
	Change24hUSD    shared.NullDecimal   `json:"change_24h_usd"`
	Visibility      asset.Visibility     `json:"visibility"`
	ValuationStatus shared.ValuationStatus `json:"valuation_status"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

type portfolioExclusionsResponse struct {
	UnpricedPositions     int  `json:"unpriced_positions"`
	NFTsExcluded          bool `json:"nfts_excluded"`
	DeFiPositionsExcluded bool `json:"defi_positions_excluded"`
}

type dataNoticeResponse struct {
	Code     portfolio.NoticeCode     `json:"code"`
	Severity portfolio.NoticeSeverity `json:"severity"`
	Params   map[string]string        `json:"params,omitempty"`
}

type portfolioResponse struct {
	WalletID              uuid.UUID                   `json:"wallet_id"`
	Currency              shared.Currency             `json:"currency"`
	TotalValueUSD         shared.NullDecimal          `json:"total_value_usd"`
	ValuationStatus       shared.ValuationStatus      `json:"valuation_status"`
	DataQuality           shared.DataQuality          `json:"data_quality"`
	DataFreshness         shared.DataFreshness        `json:"data_freshness"`
	Change24hUSD          shared.NullDecimal          `json:"change_24h_usd"`
	Change24hPct          shared.NullDecimal          `json:"change_24h_pct"`
	AsOf                  time.Time                   `json:"as_of"`
	LastSyncedAt          *time.Time                  `json:"last_synced_at"`
	CalculationVersion    int                         `json:"calculation_version"`
	Positions             []walletPositionResponse    `json:"positions"`
	VisiblePositionsCount int                         `json:"visible_positions_count"`
	HiddenPositionsCount  int                         `json:"hidden_positions_count"`
	Exclusions            portfolioExclusionsResponse `json:"exclusions"`
	Notices               []dataNoticeResponse        `json:"notices"`
}

// Get handles GET /wallets/{walletID}/portfolio.
func (h *PortfolioHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	p, err := h.service.Get(r.Context(), userID, walletID)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, r, mapPortfolio(p, h.freshness))
}

func mapPortfolio(p portfolio.Portfolio, freshness shared.FreshnessThresholds) portfolioResponse {
	positions := make([]walletPositionResponse, len(p.Positions))
	visible := 0
	hidden := 0
	for i, pos := range p.Positions {
		if pos.Visibility == asset.VisibilityVisible {
			visible++
		} else {
			hidden++
		}
		positions[i] = walletPositionResponse{
			Asset: assetResponse{
				ID:              pos.Asset.ID,
				ChainID:         pos.Asset.ChainID,
				ContractAddress: pos.Asset.ContractAddress,
				Symbol:          pos.Asset.Symbol,
				Name:            pos.Asset.Name,
				Decimals:        pos.Asset.Decimals,
				AssetType:       pos.Asset.Type,
				IconURL:         pos.Asset.IconURL,
				HasMarketData:   pos.Asset.HasMarketData(),
			},
			Balance:         pos.Balance,
			BalanceRaw:      pos.BalanceRaw,
			Price:           mapPrice(pos.Price, freshness),
			ValueUSD:        pos.ValueUSD,
			AllocationPct:   pos.AllocationPct,
			Change24hPct:    pos.Change24hPct,
			Change24hUSD:    pos.Change24hUSD,
			Visibility:      pos.Visibility,
			ValuationStatus: pos.ValuationStatus,
			UpdatedAt:       pos.UpdatedAt,
		}
	}
	if positions == nil {
		positions = []walletPositionResponse{}
	}

	notices := make([]dataNoticeResponse, len(p.Notices))
	for i, notice := range p.Notices {
		notices[i] = dataNoticeResponse{
			Code:     notice.Code,
			Severity: notice.Severity,
			Params:   notice.Params,
		}
	}
	if notices == nil {
		notices = []dataNoticeResponse{}
	}

	return portfolioResponse{
		WalletID:           p.WalletID,
		Currency:           p.Currency,
		TotalValueUSD:      p.TotalValueUSD,
		ValuationStatus:    p.ValuationStatus,
		DataQuality:        p.DataQuality,
		DataFreshness:      p.DataFreshness,
		Change24hUSD:       p.Change24hUSD,
		Change24hPct:       p.Change24hPct,
		AsOf:               p.AsOf,
		LastSyncedAt:       p.LastSyncedAt,
		CalculationVersion: p.CalculationVersion,
		Positions:          positions,
		VisiblePositionsCount: visible,
		HiddenPositionsCount:  hidden,
		Exclusions: portfolioExclusionsResponse{
			UnpricedPositions:     p.Exclusions.UnpricedPositions,
			NFTsExcluded:          p.Exclusions.NFTsExcluded,
			DeFiPositionsExcluded: p.Exclusions.DeFiPositionsExcluded,
		},
		Notices: notices,
	}
}

func mapPrice(p *price.Price, freshness shared.FreshnessThresholds) *priceResponse {
	if p == nil {
		return nil
	}
	var asOf *time.Time
	if !p.AsOf.IsZero() {
		asOf = &p.AsOf
	}
	now := time.Now().UTC()
	return &priceResponse{
		AssetID:      p.AssetID,
		ValueUSD:     p.ValueUSD,
		Currency:     p.Currency,
		Status:       p.Status,
		Freshness:    p.Freshness(freshness, now),
		AsOf:         asOf,
		Change24hPct: p.Change24h,
	}
}
