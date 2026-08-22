package scenarios

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	appportfolio "github.com/maxaicrypto/backend/internal/application/portfolio"
	appusage "github.com/maxaicrypto/backend/internal/application/usage"
	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/conversation"
	"github.com/maxaicrypto/backend/internal/domain/portfolio"
	"github.com/maxaicrypto/backend/internal/domain/provider"
	"github.com/maxaicrypto/backend/internal/domain/scenario"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/domain/subscription"
	"github.com/maxaicrypto/backend/internal/domain/usage"
)

// View is the API-facing scenario result, including an optional explanation.
type View struct {
	Result      scenario.Result
	Explanation *conversation.Response
}

// App implements Service.
type App struct {
	portfolios   appportfolio.Service
	assets       asset.Repository
	scenarios    scenario.Repository
	usage        appusage.Service
	entitlements appusage.EntitlementService
	resolver     provider.Resolver
	ai           config.AIConfig
}

// NewApp wires the scenario service.
func NewApp(
	portfolios appportfolio.Service,
	assets asset.Repository,
	scenarios scenario.Repository,
	usage appusage.Service,
	entitlements appusage.EntitlementService,
	resolver provider.Resolver,
	ai config.AIConfig,
) *App {
	return &App{
		portfolios:   portfolios,
		assets:       assets,
		scenarios:    scenarios,
		usage:        usage,
		entitlements: entitlements,
		resolver:     resolver,
		ai:           ai,
	}
}

// Simulate implements Service.
func (a *App) Simulate(ctx context.Context, userID uuid.UUID, req scenario.Request) (View, error) {
	ok, err := a.entitlements.Can(ctx, userID, subscription.FeatureScenario)
	if err != nil {
		return View{}, err
	}
	if !ok {
		return View{}, apperr.New(apperr.CodeForbidden)
	}

	reservation, err := a.usage.Reserve(ctx, userID, usage.OperationScenario, uuid.New().String())
	if err != nil {
		return View{}, err
	}

	persisted, err := a.Compute(ctx, userID, req)
	if err != nil {
		_ = a.usage.Release(ctx, reservation)
		return View{}, err
	}

	if err := a.usage.Commit(ctx, reservation); err != nil {
		return View{}, err
	}

	explanation := a.explain(ctx, persisted)
	return View{Result: persisted, Explanation: &explanation}, nil
}

// Compute implements Service.
func (a *App) Compute(ctx context.Context, userID uuid.UUID, req scenario.Request) (scenario.Result, error) {
	if err := validateRequest(req); err != nil {
		return scenario.Result{}, err
	}

	current, err := a.portfolios.Get(ctx, userID, req.WalletID)
	if err != nil {
		return scenario.Result{}, err
	}

	position, found := findPosition(current, req.AssetID)
	if !found {
		return scenario.Result{}, apperr.New(apperr.CodeValidation).
			WithMessage("The wallet does not hold this asset.").
			WithDetail("fields", map[string]string{"asset_id": "NOT_HELD"})
	}
	if !position.ValueUSD.Valid || !current.TotalValueUSD.Valid {
		return scenario.Result{}, apperr.New(apperr.CodePriceDataUnavailable).
			WithMessage("This asset cannot be simulated because its price is unknown.")
	}

	computed := computeAssetPriceChange(current, position, req.ChangePct)
	computed.UserID = userID
	computed.WalletID = req.WalletID
	computed.Type = req.Type
	computed.Currency = shared.CurrencyUSD
	computed.Asset = position.Asset
	computed.AssetID = position.Asset.ID
	computed.ChangePct = req.ChangePct
	computed.DataQuality = current.DataQuality
	computed.CalculationVersion = scenario.CalculationVersion

	persisted, err := a.scenarios.Create(ctx, computed)
	if err != nil {
		return scenario.Result{}, err
	}
	persisted.Asset = position.Asset
	persisted.CalculationID = persisted.ID
	return persisted, nil
}

// Get implements Service.
func (a *App) Get(ctx context.Context, userID, calculationID uuid.UUID) (scenario.Result, error) {
	result, err := a.scenarios.GetByID(ctx, calculationID)
	if err != nil {
		if appErr := apperr.From(err); appErr != nil && appErr.Code == apperr.CodeNotFound {
			return scenario.Result{}, apperr.New(apperr.CodeNotFound)
		}
		return scenario.Result{}, err
	}
	if result.UserID != userID {
		return scenario.Result{}, apperr.New(apperr.CodeNotFound)
	}
	ast, err := a.assets.GetByID(ctx, result.AssetID)
	if err != nil {
		return scenario.Result{}, err
	}
	result.Asset = ast
	result.CalculationID = result.ID
	return result, nil
}

func (a *App) explain(ctx context.Context, result scenario.Result) conversation.Response {
	fallback := explainScenario(result)
	if !a.ai.HasAICredentials() || a.resolver == nil {
		return fallback
	}
	llm, err := a.resolver.ResolveLLM(ctx)
	if err != nil {
		return fallback
	}
	facts, _ := json.Marshal(map[string]any{
		"symbol":                   result.Asset.Symbol,
		"change_pct":               result.ChangePct.String(),
		"baseline_portfolio_usd":   formatNull(result.Baseline.PortfolioValueUSD),
		"projected_portfolio_usd":  formatNull(result.Projection.PortfolioValueUSD),
		"portfolio_change_usd":     formatNull(result.Projection.PortfolioChangeUSD),
		"portfolio_change_pct":     formatNull(result.Projection.PortfolioChangePct),
		"asset_allocation_pct":     formatNull(result.Baseline.AssetAllocationPct),
		"asset_impact_usd":         formatNull(result.Projection.AssetImpactUSD),
		"data_quality":             result.DataQuality,
		"calculation_id":           result.CalculationID.String(),
	})
	raw, err := llm.Generate(ctx, provider.LLMRequest{
		Messages: []provider.LLMMessage{
			{
				Role: provider.RoleSystem,
				Content: "Explain the scenario using only the provided facts. Never invent numbers. Respond as JSON {\"answer\":\"...\"}.",
			},
			{Role: provider.RoleUser, Content: string(facts)},
		},
		MaxTokens: a.ai.MaxOutputTokens,
		ResponseSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"answer": map[string]any{"type": "string"},
			},
			"required":             []string{"answer"},
			"additionalProperties": false,
		},
	})
	if err != nil {
		return fallback
	}
	var payload struct {
		Answer string `json:"answer"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw.Content)), &payload); err != nil || strings.TrimSpace(payload.Answer) == "" {
		return fallback
	}
	fallback.Answer = strings.TrimSpace(payload.Answer)
	if len(fallback.Claims) > 0 {
		fallback.Claims[0].Text = fallback.Answer
	}
	return fallback
}

func validateRequest(req scenario.Request) error {
	if req.WalletID == uuid.Nil {
		return apperr.New(apperr.CodeScenarioInvalid).
			WithDetail("fields", map[string]string{"wallet_id": "is required"})
	}
	if req.AssetID == uuid.Nil {
		return apperr.New(apperr.CodeScenarioInvalid).
			WithDetail("fields", map[string]string{"asset_id": "is required"})
	}
	if req.Type != scenario.TypeAssetPriceChange {
		return apperr.New(apperr.CodeScenarioInvalid).
			WithMessage("Only ASSET_PRICE_CHANGE scenarios are supported.").
			WithDetail("fields", map[string]string{"type": "must be ASSET_PRICE_CHANGE"})
	}
	return nil
}

func findPosition(p portfolio.Portfolio, assetID uuid.UUID) (portfolio.Position, bool) {
	for _, pos := range p.Positions {
		if pos.Asset.ID == assetID {
			return pos, true
		}
	}
	return portfolio.Position{}, false
}

func computeAssetPriceChange(
	current portfolio.Portfolio,
	position portfolio.Position,
	changePct shared.Decimal,
) scenario.Result {
	hundred := decimal.NewFromInt(100)
	factor := decimal.NewFromInt(1).Add(changePct.Value().Div(hundred))
	assetValue := position.ValueUSD.Decimal.Value()
	portfolioValue := current.TotalValueUSD.Decimal.Value()

	projectedAsset := assetValue.Mul(factor)
	impact := projectedAsset.Sub(assetValue)
	projectedPortfolio := portfolioValue.Add(impact)
	changePortfolioPct := impact.Div(portfolioValue).Mul(hundred)

	return scenario.Result{
		Baseline: scenario.Baseline{
			PortfolioValueUSD:  current.TotalValueUSD,
			AssetValueUSD:      position.ValueUSD,
			AssetAllocationPct: position.AllocationPct,
		},
		Projection: scenario.Projection{
			PortfolioValueUSD:  shared.Known(shared.NewDecimal(projectedPortfolio)),
			AssetValueUSD:      shared.Known(shared.NewDecimal(projectedAsset)),
			AssetImpactUSD:     shared.Known(shared.NewDecimal(impact)),
			PortfolioChangeUSD: shared.Known(shared.NewDecimal(impact)),
			PortfolioChangePct: shared.Known(shared.NewDecimal(changePortfolioPct)),
		},
	}
}

func explainScenario(result scenario.Result) conversation.Response {
	symbol := result.Asset.Symbol
	change := result.ChangePct.String()
	from := formatNull(result.Baseline.PortfolioValueUSD)
	to := formatNull(result.Projection.PortfolioValueUSD)
	delta := formatNull(result.Projection.PortfolioChangeUSD)
	deltaPct := formatNull(result.Projection.PortfolioChangePct)
	allocation := formatNull(result.Baseline.AssetAllocationPct)
	impact := formatNull(result.Projection.AssetImpactUSD)

	answer := fmt.Sprintf(
		"If %s moves %s%%, your portfolio goes from $%s to $%s — a change of $%s (%s%%). %s is %s%% of your portfolio, so the position itself changes by $%s.",
		symbol, change, from, to, delta, deltaPct, symbol, allocation, impact,
	)

	return conversation.Response{
		Answer:      answer,
		Intent:      conversation.IntentScenarioSimulation,
		DataQuality: result.DataQuality,
		Claims: []conversation.Claim{
			{
				Text: answer,
				Evidence: []conversation.Evidence{
					{Type: conversation.EvidenceScenario, ID: result.CalculationID.String()},
				},
			},
		},
		References: []conversation.Reference{
			{Type: conversation.ReferenceAsset, ID: result.Asset.ID.String(), Label: &symbol},
			{Type: conversation.ReferenceScenario, ID: result.ID.String()},
		},
	}
}

func formatNull(value shared.NullDecimal) string {
	if !value.Valid {
		return "—"
	}
	return strings.TrimRight(strings.TrimRight(value.Decimal.Value().StringFixed(2), "0"), ".")
}

var _ Service = (*App)(nil)
