package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	appportfolio "github.com/maxaicrypto/backend/internal/application/portfolio"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/performance"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/transport/http/middleware"
	"github.com/maxaicrypto/backend/internal/transport/http/request"
	"github.com/maxaicrypto/backend/internal/transport/http/response"
)

// PerformanceHandler serves wallet performance routes.
type PerformanceHandler struct {
	service appportfolio.PerformanceService
}

// NewPerformanceHandler builds the performance handler.
func NewPerformanceHandler(service appportfolio.PerformanceService) *PerformanceHandler {
	return &PerformanceHandler{service: service}
}

type performanceEndpointResponse struct {
	CapturedAt time.Time              `json:"captured_at"`
	ValueUSD   shared.Decimal         `json:"value_usd"`
	Status     shared.ValuationStatus `json:"status"`
}

type performanceSnapshotPointResponse struct {
	CapturedAt    time.Time              `json:"captured_at"`
	TotalValueUSD shared.NullDecimal     `json:"total_value_usd"`
	Status        shared.ValuationStatus `json:"status"`
}

type performanceDriverResponse struct {
	Asset           assetResponse      `json:"asset"`
	AllocationPct   shared.NullDecimal `json:"allocation_pct"`
	ContributionUSD shared.NullDecimal `json:"contribution_usd"`
	ContributionPct shared.NullDecimal `json:"contribution_pct"`
	ChangePct       shared.NullDecimal `json:"change_pct"`
}

type portfolioPerformanceResponse struct {
	WalletID           uuid.UUID                          `json:"wallet_id"`
	Period             performance.Period                 `json:"period"`
	Status             performance.Status                 `json:"status"`
	DataQuality        shared.DataQuality                 `json:"data_quality"`
	Currency           shared.Currency                    `json:"currency"`
	Opening            *performanceEndpointResponse       `json:"opening"`
	Closing            *performanceEndpointResponse       `json:"closing"`
	ChangeUSD          shared.NullDecimal                 `json:"change_usd"`
	ChangePct          shared.NullDecimal                 `json:"change_pct"`
	Series             []performanceSnapshotPointResponse `json:"series"`
	Drivers            []performanceDriverResponse        `json:"drivers"`
	UnavailableReason  *string                            `json:"unavailable_reason"`
	CalculationID      *uuid.UUID                         `json:"calculation_id"`
	CalculationVersion *int                               `json:"calculation_version"`
}

// Get handles GET /wallets/{walletID}/performance.
func (h *PerformanceHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	rawPeriod, ok := request.QueryString(r, "period")
	if !ok {
		response.Error(w, r, apperr.New(apperr.CodeValidation).
			WithMessage("The period parameter is required.").
			WithDetail("fields", map[string]string{"period": "is required"}))
		return
	}
	period, ok := performance.ParsePeriod(rawPeriod)
	if !ok {
		response.Error(w, r, apperr.New(apperr.CodeValidation).
			WithMessage("The period parameter is not valid.").
			WithDetail("fields", map[string]string{"period": "must be one of 24h, 7d, 30d, all"}))
		return
	}

	result, err := h.service.Get(r.Context(), userID, walletID, period)
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, r, mapPerformance(result))
}

func mapPerformance(p performance.Performance) portfolioPerformanceResponse {
	series := make([]performanceSnapshotPointResponse, len(p.Series))
	for i, point := range p.Series {
		series[i] = performanceSnapshotPointResponse{
			CapturedAt:    point.CapturedAt,
			TotalValueUSD: point.TotalValueUSD,
			Status:        point.Status,
		}
	}
	if series == nil {
		series = []performanceSnapshotPointResponse{}
	}

	drivers := make([]performanceDriverResponse, len(p.Drivers))
	for i, driver := range p.Drivers {
		drivers[i] = performanceDriverResponse{
			Asset: assetResponse{
				ID:              driver.Asset.ID,
				ChainID:         driver.Asset.ChainID,
				ContractAddress: driver.Asset.ContractAddress,
				Symbol:          driver.Asset.Symbol,
				Name:            driver.Asset.Name,
				Decimals:        driver.Asset.Decimals,
				AssetType:       driver.Asset.Type,
				IconURL:         driver.Asset.IconURL,
				HasMarketData:   driver.Asset.HasMarketData(),
			},
			AllocationPct:   driver.AllocationPct,
			ContributionUSD: driver.ContributionUSD,
			ContributionPct: driver.ContributionPct,
			ChangePct:       driver.ChangePct,
		}
	}
	if drivers == nil {
		drivers = []performanceDriverResponse{}
	}

	return portfolioPerformanceResponse{
		WalletID:           p.WalletID,
		Period:             p.Period,
		Status:             p.Status,
		DataQuality:        p.DataQuality,
		Currency:           p.Currency,
		Opening:            mapPerformanceEndpoint(p.Opening),
		Closing:            mapPerformanceEndpoint(p.Closing),
		ChangeUSD:          p.ChangeUSD,
		ChangePct:          p.ChangePct,
		Series:             series,
		Drivers:            drivers,
		UnavailableReason:  p.UnavailableReason,
		CalculationID:      p.CalculationID,
		CalculationVersion: p.CalculationVersion,
	}
}

func mapPerformanceEndpoint(endpoint *performance.Endpoint) *performanceEndpointResponse {
	if endpoint == nil {
		return nil
	}
	return &performanceEndpointResponse{
		CapturedAt: endpoint.CapturedAt,
		ValueUSD:   endpoint.ValueUSD,
		Status:     endpoint.Status,
	}
}
