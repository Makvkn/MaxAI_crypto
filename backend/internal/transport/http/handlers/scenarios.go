package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	appscenarios "github.com/maxaicrypto/backend/internal/application/scenarios"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/scenario"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/transport/http/middleware"
	"github.com/maxaicrypto/backend/internal/transport/http/request"
	"github.com/maxaicrypto/backend/internal/transport/http/response"
)

type scenarioRequestBody struct {
	WalletID  uuid.UUID       `json:"wallet_id"`
	Type      scenario.Type   `json:"type"`
	AssetID   uuid.UUID       `json:"asset_id"`
	ChangePct shared.Decimal  `json:"change_pct"`
}

type scenarioBaselineResponse struct {
	PortfolioValueUSD  shared.NullDecimal `json:"portfolio_value_usd"`
	AssetValueUSD      shared.NullDecimal `json:"asset_value_usd"`
	AssetAllocationPct shared.NullDecimal `json:"asset_allocation_pct"`
}

type scenarioProjectionResponse struct {
	PortfolioValueUSD  shared.NullDecimal `json:"portfolio_value_usd"`
	AssetValueUSD      shared.NullDecimal `json:"asset_value_usd"`
	AssetImpactUSD     shared.NullDecimal `json:"asset_impact_usd"`
	PortfolioChangeUSD shared.NullDecimal `json:"portfolio_change_usd"`
	PortfolioChangePct shared.NullDecimal `json:"portfolio_change_pct"`
}

type scenarioResultResponse struct {
	ID                 uuid.UUID                   `json:"id"`
	WalletID           uuid.UUID                   `json:"wallet_id"`
	Type               scenario.Type               `json:"type"`
	Currency           shared.Currency             `json:"currency"`
	Asset              assetResponse               `json:"asset"`
	ChangePct          shared.Decimal              `json:"change_pct"`
	Baseline           scenarioBaselineResponse    `json:"baseline"`
	Projection         scenarioProjectionResponse  `json:"projection"`
	DataQuality        shared.DataQuality          `json:"data_quality"`
	CalculationID      uuid.UUID                   `json:"calculation_id"`
	CalculationVersion int                         `json:"calculation_version"`
	CreatedAt          time.Time                   `json:"created_at"`
	Explanation        *aiResponseBody             `json:"explanation"`
}

// SimulateScenario handles POST /ai/scenarios.
func (h *AIHandler) SimulateScenario(w http.ResponseWriter, r *http.Request) {
	if h.scenarios == nil {
		response.Error(w, r, apperr.New(apperr.CodeNotImplemented))
		return
	}
	userID, ok := middleware.UserIDFrom(r.Context())
	if !ok {
		response.Error(w, r, apperr.New(apperr.CodeAuthentication))
		return
	}

	var body scenarioRequestBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.Error(w, r, err)
		return
	}

	view, err := h.scenarios.Simulate(r.Context(), userID, scenario.Request{
		WalletID:  body.WalletID,
		Type:      body.Type,
		AssetID:   body.AssetID,
		ChangePct: body.ChangePct,
	})
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, r, mapScenarioView(view))
}

func mapScenarioView(view appscenarios.View) scenarioResultResponse {
	result := view.Result
	return scenarioResultResponse{
		ID:       result.ID,
		WalletID: result.WalletID,
		Type:     result.Type,
		Currency: result.Currency,
		Asset: assetResponse{
			ID:              result.Asset.ID,
			ChainID:         result.Asset.ChainID,
			ContractAddress: result.Asset.ContractAddress,
			Symbol:          result.Asset.Symbol,
			Name:            result.Asset.Name,
			Decimals:        result.Asset.Decimals,
			AssetType:       result.Asset.Type,
			IconURL:         result.Asset.IconURL,
			HasMarketData:   result.Asset.HasMarketData(),
		},
		ChangePct: result.ChangePct,
		Baseline: scenarioBaselineResponse{
			PortfolioValueUSD:  result.Baseline.PortfolioValueUSD,
			AssetValueUSD:      result.Baseline.AssetValueUSD,
			AssetAllocationPct: result.Baseline.AssetAllocationPct,
		},
		Projection: scenarioProjectionResponse{
			PortfolioValueUSD:  result.Projection.PortfolioValueUSD,
			AssetValueUSD:      result.Projection.AssetValueUSD,
			AssetImpactUSD:     result.Projection.AssetImpactUSD,
			PortfolioChangeUSD: result.Projection.PortfolioChangeUSD,
			PortfolioChangePct: result.Projection.PortfolioChangePct,
		},
		DataQuality:        result.DataQuality,
		CalculationID:      result.CalculationID,
		CalculationVersion: result.CalculationVersion,
		CreatedAt:          result.CreatedAt,
		Explanation:        mapAIResponse(view.Explanation),
	}
}
