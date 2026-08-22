package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	appportfolio "github.com/maxaicrypto/backend/internal/application/portfolio"
	apppricing "github.com/maxaicrypto/backend/internal/application/pricing"
	appscenarios "github.com/maxaicrypto/backend/internal/application/scenarios"
	apptransactions "github.com/maxaicrypto/backend/internal/application/transactions"
	"github.com/maxaicrypto/backend/internal/domain/conversation"
	"github.com/maxaicrypto/backend/internal/domain/performance"
	"github.com/maxaicrypto/backend/internal/domain/portfolio"
	"github.com/maxaicrypto/backend/internal/domain/scenario"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/domain/transaction"

	"github.com/google/uuid"
)

// portfolioTool loads the current portfolio facts for the model.
type portfolioTool struct {
	portfolios appportfolio.Service
}

func newPortfolioTool(portfolios appportfolio.Service) *portfolioTool {
	return &portfolioTool{portfolios: portfolios}
}

func (t *portfolioTool) Name() conversation.ToolName { return conversation.ToolGetPortfolio }

func (t *portfolioTool) Description() string {
	return "Returns the current wallet portfolio valuation, allocation and data quality."
}

func (t *portfolioTool) ParametersSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{}}
}

func (t *portfolioTool) Execute(ctx context.Context, invocation ToolInvocation) ([]byte, error) {
	p, err := t.portfolios.Get(ctx, invocation.UserID, invocation.WalletID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(mapPortfolioFacts(p))
}

type positionsTool struct {
	portfolios appportfolio.Service
}

func newPositionsTool(portfolios appportfolio.Service) *positionsTool {
	return &positionsTool{portfolios: portfolios}
}

func (t *positionsTool) Name() conversation.ToolName { return conversation.ToolGetPositions }

func (t *positionsTool) Description() string {
	return "Returns the wallet's current positions with balances, values and visibility."
}

func (t *positionsTool) ParametersSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{}}
}

func (t *positionsTool) Execute(ctx context.Context, invocation ToolInvocation) ([]byte, error) {
	p, err := t.portfolios.Get(ctx, invocation.UserID, invocation.WalletID)
	if err != nil {
		return nil, err
	}
	facts := mapPortfolioFacts(p)
	return json.Marshal(map[string]any{
		"wallet_id":    facts["wallet_id"],
		"data_quality": facts["data_quality"],
		"positions":    facts["positions"],
	})
}

type performanceTool struct {
	performance appportfolio.PerformanceService
}

func newPerformanceTool(performance appportfolio.PerformanceService) *performanceTool {
	return &performanceTool{performance: performance}
}

func (t *performanceTool) Name() conversation.ToolName {
	return conversation.ToolGetPortfolioPerformance
}

func (t *performanceTool) Description() string {
	return "Returns snapshot-based portfolio performance for a supported period."
}

func (t *performanceTool) ParametersSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"period": map[string]any{
				"type": "string",
				"enum": []string{"24h", "7d", "30d", "all"},
			},
		},
		"required": []string{"period"},
	}
}

func (t *performanceTool) Execute(ctx context.Context, invocation ToolInvocation) ([]byte, error) {
	period := performance.Period24h
	var args struct {
		Period string `json:"period"`
	}
	if len(invocation.Arguments) > 0 {
		if err := json.Unmarshal(invocation.Arguments, &args); err != nil {
			return nil, err
		}
		if parsed, ok := performance.ParsePeriod(args.Period); ok {
			period = parsed
		}
	}
	result, err := t.performance.Get(ctx, invocation.UserID, invocation.WalletID, period)
	if err != nil {
		return nil, err
	}
	return json.Marshal(mapPerformanceFacts(result))
}

type transactionTool struct {
	transactions apptransactions.Service
}

func newTransactionTool(transactions apptransactions.Service) *transactionTool {
	return &transactionTool{transactions: transactions}
}

func (t *transactionTool) Name() conversation.ToolName { return conversation.ToolGetTransaction }

func (t *transactionTool) Description() string {
	return "Returns one canonical wallet transaction with amounts, assets and classification."
}

func (t *transactionTool) ParametersSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"transaction_id": map[string]any{"type": "string", "format": "uuid"},
		},
		"required": []string{"transaction_id"},
	}
}

func (t *transactionTool) Execute(ctx context.Context, invocation ToolInvocation) ([]byte, error) {
	var args struct {
		TransactionID string `json:"transaction_id"`
	}
	if err := json.Unmarshal(invocation.Arguments, &args); err != nil {
		return nil, err
	}
	txID, err := uuid.Parse(args.TransactionID)
	if err != nil {
		return nil, err
	}
	view, err := t.transactions.Get(ctx, invocation.UserID, invocation.WalletID, txID)
	if err != nil {
		return nil, err
	}
	return json.Marshal(mapTransactionFacts(view))
}

type historicalTool struct {
	portfolios appportfolio.Service
	snapshots  portfolio.SnapshotRepository
}

func newHistoricalTool(portfolios appportfolio.Service, snapshots portfolio.SnapshotRepository) *historicalTool {
	return &historicalTool{portfolios: portfolios, snapshots: snapshots}
}

func (t *historicalTool) Name() conversation.ToolName {
	return conversation.ToolGetHistoricalPortfolio
}

func (t *historicalTool) Description() string {
	return "Returns recent historical portfolio snapshot totals for the wallet."
}

func (t *historicalTool) ParametersSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{}}
}

func (t *historicalTool) Execute(ctx context.Context, invocation ToolInvocation) ([]byte, error) {
	if _, err := t.portfolios.Get(ctx, invocation.UserID, invocation.WalletID); err != nil {
		return nil, err
	}
	to := time.Now().UTC()
	from := to.Add(-30 * 24 * time.Hour)
	rows, err := t.snapshots.ListBetween(ctx, invocation.WalletID, from, to, 30)
	if err != nil {
		return nil, err
	}
	points := make([]map[string]any, 0, len(rows))
	for _, snap := range rows {
		point := map[string]any{
			"snapshot_id":  snap.ID.String(),
			"captured_at":  snap.CapturedAt.Format(time.RFC3339),
			"status":       snap.Status,
			"data_quality": snap.DataQuality,
		}
		if snap.TotalValueUSD.Valid {
			point["total_value_usd"] = snap.TotalValueUSD.Decimal.String()
		}
		points = append(points, point)
	}
	quality := shared.DataQualityComplete
	if len(points) == 0 {
		quality = shared.DataQualityUnavailable
	}
	return json.Marshal(map[string]any{
		"wallet_id":    invocation.WalletID.String(),
		"data_quality": quality,
		"snapshots":    points,
	})
}

type assetPriceTool struct {
	portfolios appportfolio.Service
	pricing    apppricing.Service
}

func newAssetPriceTool(portfolios appportfolio.Service, pricing apppricing.Service) *assetPriceTool {
	return &assetPriceTool{portfolios: portfolios, pricing: pricing}
}

func (t *assetPriceTool) Name() conversation.ToolName { return conversation.ToolGetAssetPrice }

func (t *assetPriceTool) Description() string {
	return "Returns the current USD price for one asset held in the wallet portfolio."
}

func (t *assetPriceTool) ParametersSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"asset_id": map[string]any{"type": "string", "format": "uuid"},
		},
		"required": []string{"asset_id"},
	}
}

func (t *assetPriceTool) Execute(ctx context.Context, invocation ToolInvocation) ([]byte, error) {
	var args struct {
		AssetID string `json:"asset_id"`
	}
	if err := json.Unmarshal(invocation.Arguments, &args); err != nil {
		return nil, err
	}
	assetID, err := uuid.Parse(args.AssetID)
	if err != nil {
		return nil, err
	}
	p, err := t.portfolios.Get(ctx, invocation.UserID, invocation.WalletID)
	if err != nil {
		return nil, err
	}
	var symbol, name string
	held := false
	for _, pos := range p.Positions {
		if pos.Asset.ID == assetID {
			held = true
			symbol = pos.Asset.Symbol
			name = pos.Asset.Name
			break
		}
	}
	if !held {
		return json.Marshal(map[string]any{
			"asset_id":     assetID.String(),
			"held":         false,
			"data_quality": shared.DataQualityUnavailable,
		})
	}
	prices, err := t.pricing.GetCurrent(ctx, []uuid.UUID{assetID})
	if err != nil {
		return nil, err
	}
	facts := map[string]any{
		"asset_id":     assetID.String(),
		"symbol":       symbol,
		"name":         name,
		"held":         true,
		"data_quality": shared.DataQualityUnavailable,
	}
	if quote, ok := prices[assetID]; ok && quote.IsUsable() {
		facts["price_usd"] = quote.ValueUSD.Decimal.String()
		facts["as_of"] = quote.AsOf.Format(time.RFC3339)
		facts["data_quality"] = shared.DataQualityComplete
		if quote.Change24h.Valid {
			facts["change_24h_pct"] = quote.Change24h.Decimal.String()
		}
	}
	return json.Marshal(facts)
}

type scenarioTool struct {
	scenarios appscenarios.Service
}

func newScenarioTool(scenarios appscenarios.Service) *scenarioTool {
	return &scenarioTool{scenarios: scenarios}
}

func (t *scenarioTool) Name() conversation.ToolName { return conversation.ToolSimulateScenario }

func (t *scenarioTool) Description() string {
	return "Runs a deterministic ASSET_PRICE_CHANGE scenario for an asset held by the wallet."
}

func (t *scenarioTool) ParametersSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"asset_id":   map[string]any{"type": "string", "format": "uuid"},
			"change_pct": map[string]any{"type": "string"},
		},
		"required": []string{"asset_id", "change_pct"},
	}
}

func (t *scenarioTool) Execute(ctx context.Context, invocation ToolInvocation) ([]byte, error) {
	var args struct {
		AssetID   string `json:"asset_id"`
		ChangePct string `json:"change_pct"`
	}
	if err := json.Unmarshal(invocation.Arguments, &args); err != nil {
		return nil, err
	}
	assetID, err := uuid.Parse(args.AssetID)
	if err != nil {
		return nil, err
	}
	change, err := shared.ParseDecimal(args.ChangePct)
	if err != nil {
		return nil, err
	}
	result, err := t.scenarios.Compute(ctx, invocation.UserID, scenario.Request{
		WalletID:  invocation.WalletID,
		AssetID:   assetID,
		Type:      scenario.TypeAssetPriceChange,
		ChangePct: change,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(mapScenarioFacts(result))
}

type registry struct {
	tools map[conversation.ToolName]Tool
}

func newRegistry(tools ...Tool) *registry {
	index := make(map[conversation.ToolName]Tool, len(tools))
	for _, tool := range tools {
		index[tool.Name()] = tool
	}
	return &registry{tools: index}
}

func (r *registry) Get(name conversation.ToolName) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

func (r *registry) All() []Tool {
	items := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		items = append(items, tool)
	}
	return items
}

var _ ToolRegistry = (*registry)(nil)

func mapPortfolioFacts(p portfolio.Portfolio) map[string]any {
	positions := make([]map[string]any, 0, len(p.Positions))
	for _, pos := range p.Positions {
		item := map[string]any{
			"asset_id":   pos.Asset.ID.String(),
			"symbol":     pos.Asset.Symbol,
			"balance":    pos.Balance.String(),
			"visibility": pos.Visibility,
		}
		if pos.ValueUSD.Valid {
			item["value_usd"] = pos.ValueUSD.Decimal.String()
		}
		if pos.AllocationPct.Valid {
			item["allocation_pct"] = pos.AllocationPct.Decimal.String()
		}
		positions = append(positions, item)
	}
	facts := map[string]any{
		"wallet_id":        p.WalletID.String(),
		"valuation_status": p.ValuationStatus,
		"data_quality":     p.DataQuality,
		"positions":        positions,
	}
	if p.TotalValueUSD.Valid {
		facts["total_value_usd"] = p.TotalValueUSD.Decimal.String()
	}
	return facts
}

func mapPerformanceFacts(p performance.Performance) map[string]any {
	facts := map[string]any{
		"wallet_id":    p.WalletID.String(),
		"period":       p.Period,
		"status":       p.Status,
		"data_quality": p.DataQuality,
	}
	if p.ChangePct.Valid {
		facts["change_pct"] = p.ChangePct.Decimal.String()
	}
	if p.ChangeUSD.Valid {
		facts["change_usd"] = p.ChangeUSD.Decimal.String()
	}
	return facts
}

func mapTransactionFacts(view apptransactions.View) map[string]any {
	tx := view.Transaction
	facts := map[string]any{
		"transaction_id": tx.ID.String(),
		"wallet_id":      tx.WalletID.String(),
		"chain_id":       tx.ChainID,
		"tx_hash":        tx.TxHash,
		"timestamp":      tx.Timestamp.Format(time.RFC3339),
		"status":         tx.Status,
		"type":           tx.Type,
		"data_quality":   shared.DataQualityComplete,
	}
	if tx.FromAddress != nil {
		facts["from_address"] = *tx.FromAddress
	}
	if tx.ToAddress != nil {
		facts["to_address"] = *tx.ToAddress
	}
	if tx.Protocol != nil {
		facts["protocol"] = *tx.Protocol
	}
	if view.AssetIn != nil {
		facts["asset_in"] = map[string]any{"id": view.AssetIn.ID.String(), "symbol": view.AssetIn.Symbol}
	}
	if view.AssetOut != nil {
		facts["asset_out"] = map[string]any{"id": view.AssetOut.ID.String(), "symbol": view.AssetOut.Symbol}
	}
	if tx.AmountIn.Valid {
		facts["amount_in"] = tx.AmountIn.Decimal.String()
	}
	if tx.AmountOut.Valid {
		facts["amount_out"] = tx.AmountOut.Decimal.String()
	}
	if view.ValueInUSD.Valid {
		facts["value_in_usd"] = view.ValueInUSD.Decimal.String()
	}
	if view.ValueOutUSD.Valid {
		facts["value_out_usd"] = view.ValueOutUSD.Decimal.String()
	}
	if view.ExplorerURL != nil {
		facts["explorer_url"] = *view.ExplorerURL
	}
	if tx.Type == transaction.TypeUnknown {
		facts["data_quality"] = shared.DataQualityPartial
	}
	return facts
}

func mapScenarioFacts(result scenario.Result) map[string]any {
	facts := map[string]any{
		"calculation_id": result.CalculationID.String(),
		"wallet_id":      result.WalletID.String(),
		"type":           result.Type,
		"asset_id":       result.AssetID.String(),
		"symbol":         result.Asset.Symbol,
		"change_pct":     result.ChangePct.String(),
		"data_quality":   result.DataQuality,
	}
	if result.Baseline.PortfolioValueUSD.Valid {
		facts["baseline_portfolio_usd"] = result.Baseline.PortfolioValueUSD.Decimal.String()
	}
	if result.Projection.PortfolioValueUSD.Valid {
		facts["projected_portfolio_usd"] = result.Projection.PortfolioValueUSD.Decimal.String()
	}
	if result.Projection.PortfolioChangeUSD.Valid {
		facts["portfolio_change_usd"] = result.Projection.PortfolioChangeUSD.Decimal.String()
	}
	if result.Projection.PortfolioChangePct.Valid {
		facts["portfolio_change_pct"] = result.Projection.PortfolioChangePct.Decimal.String()
	}
	return facts
}

func resolveIntent(question string, req OrchestrationRequest) conversation.Intent {
	if req.ScenarioID != nil {
		return conversation.IntentScenarioSimulation
	}
	if req.TransactionID != nil {
		return conversation.IntentTransactionExplain
	}

	text := strings.ToLower(question)
	switch {
	case containsAny(text, "news", "twitter", "regulation", "hack", "listing"):
		return conversation.IntentUnsupported
	case containsAny(text, "should i", "buy", "sell", "short", "long", "leverage"):
		return conversation.IntentUnsupported
	case containsAny(text, "scenario", "what if", "if eth", "if btc", "simulate"):
		return conversation.IntentScenarioSimulation
	case containsAny(text, "allocation", "largest", "concentration", "diversif"):
		return conversation.IntentPortfolioAllocation
	case containsAny(text, "performance", "week", "month", "down", "up", "24h", "7d", "30d"):
		return conversation.IntentPortfolioPerformance
	case containsAny(text, "transaction", "transfer", "swap", "tx"):
		return conversation.IntentTransactionExplain
	case containsAny(text, "history", "historical", "snapshot"):
		return conversation.IntentGeneralQuestion
	default:
		return conversation.IntentPortfolioSummary
	}
}

func containsAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

func streamText(ctx context.Context, sink StreamSink, text string) error {
	chunkSize := 48
	runes := []rune(text)
	for start := 0; start < len(runes); start += chunkSize {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		if err := sink.Send(ctx, Event{Type: EventTextDelta, Text: string(runes[start:end])}); err != nil {
			return err
		}
	}
	return nil
}

func runTool(
	ctx context.Context,
	tool Tool,
	invocation ToolInvocation,
	sink StreamSink,
) ([]byte, conversation.ToolCall, error) {
	callID := fmt.Sprintf("tool_%d", time.Now().UTC().UnixNano())
	call := conversation.ToolCall{
		ID:        callID,
		Tool:      tool.Name(),
		Status:    conversation.ToolCallRunning,
		StartedAt: time.Now().UTC(),
	}
	if err := sink.Send(ctx, Event{
		Type:       EventToolStarted,
		ToolCallID: callID,
		Tool:       tool.Name(),
	}); err != nil {
		return nil, call, err
	}

	result, err := tool.Execute(ctx, invocation)
	now := time.Now().UTC()
	call.CompletedAt = &now
	if err != nil {
		call.Status = conversation.ToolCallFailed
		_ = sink.Send(ctx, Event{
			Type:       EventToolCompleted,
			ToolCallID: callID,
			Tool:       tool.Name(),
			ToolOK:     false,
		})
		return nil, call, err
	}
	call.Status = conversation.ToolCallCompleted
	if err := sink.Send(ctx, Event{
		Type:       EventToolCompleted,
		ToolCallID: callID,
		Tool:       tool.Name(),
		ToolOK:     true,
	}); err != nil {
		return result, call, err
	}
	return result, call, nil
}

func portfolioSummaryAnswer(facts map[string]any, quality shared.DataQuality) conversation.Response {
	total, _ := facts["total_value_usd"].(string)
	walletID, _ := facts["wallet_id"].(string)
	answer := "I reviewed your current portfolio snapshot."
	if total != "" {
		answer = fmt.Sprintf("Your portfolio is currently valued at $%s.", total)
	}
	claims := []conversation.Claim{}
	if total != "" {
		claims = append(claims, conversation.Claim{
			Text: fmt.Sprintf("Portfolio total value is $%s", total),
			Evidence: []conversation.Evidence{
				{Type: conversation.EvidencePortfolio, ID: walletID},
			},
		})
	}
	return conversation.Response{
		Answer:      answer,
		Intent:      conversation.IntentPortfolioSummary,
		DataQuality: quality,
		Claims:      claims,
		References: []conversation.Reference{{
			Type: conversation.ReferencePortfolio,
			ID:   walletID,
		}},
	}
}

func portfolioPerformanceAnswer(facts map[string]any, quality shared.DataQuality) conversation.Response {
	status, _ := facts["status"].(string)
	walletID, _ := facts["wallet_id"].(string)
	if performance.Status(status) == performance.StatusUnavailable {
		reason := "Performance data is unavailable for the requested period."
		return conversation.Response{
			Answer:            "I couldn't compute portfolio performance because historical snapshots are not available yet.",
			Intent:            conversation.IntentPortfolioPerformance,
			DataQuality:       shared.DataQualityUnavailable,
			UnsupportedReason: &reason,
			Claims:            []conversation.Claim{},
			References:        []conversation.Reference{},
		}
	}
	changePct, _ := facts["change_pct"].(string)
	period, _ := facts["period"].(string)
	answer := fmt.Sprintf("Portfolio performance for %s is unavailable.", period)
	if changePct != "" {
		answer = fmt.Sprintf("Portfolio performance for %s is %s%%.", period, changePct)
	}
	claims := []conversation.Claim{}
	if changePct != "" {
		claims = append(claims, conversation.Claim{
			Text: fmt.Sprintf("Performance for %s is %s%%", period, changePct),
			Evidence: []conversation.Evidence{
				{Type: conversation.EvidencePortfolioPerformace, ID: walletID},
			},
		})
	}
	return conversation.Response{
		Answer:      answer,
		Intent:      conversation.IntentPortfolioPerformance,
		DataQuality: quality,
		Claims:      claims,
		References: []conversation.Reference{{
			Type: conversation.ReferencePortfolio,
			ID:   walletID,
		}},
	}
}

func transactionExplainAnswer(facts map[string]any, quality shared.DataQuality) conversation.Response {
	txType, _ := facts["type"].(string)
	txID, _ := facts["transaction_id"].(string)
	hash, _ := facts["tx_hash"].(string)
	answer := fmt.Sprintf("This transaction is classified as %s.", txType)
	if hash != "" {
		answer = fmt.Sprintf("Transaction %s is classified as %s.", shortHash(hash), txType)
	}
	if in, ok := facts["asset_in"].(map[string]any); ok {
		if symbol, _ := in["symbol"].(string); symbol != "" {
			if amount, _ := facts["amount_in"].(string); amount != "" {
				answer += fmt.Sprintf(" You received %s %s.", amount, symbol)
			}
		}
	}
	if out, ok := facts["asset_out"].(map[string]any); ok {
		if symbol, _ := out["symbol"].(string); symbol != "" {
			if amount, _ := facts["amount_out"].(string); amount != "" {
				answer += fmt.Sprintf(" You sent %s %s.", amount, symbol)
			}
		}
	}
	return conversation.Response{
		Answer:      answer,
		Intent:      conversation.IntentTransactionExplain,
		DataQuality: quality,
		Claims: []conversation.Claim{{
			Text: answer,
			Evidence: []conversation.Evidence{
				{Type: conversation.EvidenceTransaction, ID: txID},
			},
		}},
		References: []conversation.Reference{{
			Type: conversation.ReferenceTransaction,
			ID:   txID,
		}},
	}
}

func scenarioExplainAnswer(facts map[string]any, quality shared.DataQuality) conversation.Response {
	symbol, _ := facts["symbol"].(string)
	change, _ := facts["change_pct"].(string)
	from, _ := facts["baseline_portfolio_usd"].(string)
	to, _ := facts["projected_portfolio_usd"].(string)
	delta, _ := facts["portfolio_change_usd"].(string)
	deltaPct, _ := facts["portfolio_change_pct"].(string)
	calcID, _ := facts["calculation_id"].(string)
	answer := fmt.Sprintf(
		"If %s moves %s%%, your portfolio goes from $%s to $%s — a change of $%s (%s%%).",
		symbol, change, from, to, delta, deltaPct,
	)
	return conversation.Response{
		Answer:      answer,
		Intent:      conversation.IntentScenarioSimulation,
		DataQuality: quality,
		Claims: []conversation.Claim{{
			Text: answer,
			Evidence: []conversation.Evidence{
				{Type: conversation.EvidenceScenario, ID: calcID},
			},
		}},
		References: []conversation.Reference{{
			Type: conversation.ReferenceScenario,
			ID:   calcID,
		}},
	}
}

func unsupportedAnswer() conversation.Response {
	reason := "This question is outside the MVP scope."
	return conversation.Response{
		Answer:            "I can explain your portfolio, performance, transactions and scenarios, but I can't answer news or trading advice questions yet.",
		Intent:            conversation.IntentUnsupported,
		DataQuality:       shared.DataQualityComplete,
		UnsupportedReason: &reason,
		Claims:            []conversation.Claim{},
		References:        []conversation.Reference{},
	}
}

func shortHash(hash string) string {
	if len(hash) <= 12 {
		return hash
	}
	return hash[:10] + "…"
}

func missingTransactionAnswer() conversation.Response {
	reason := "A specific transaction_id is required to explain a transaction."
	return conversation.Response{
		Answer:            "Please open a specific transaction so I can explain it with backend facts.",
		Intent:            conversation.IntentTransactionExplain,
		DataQuality:       shared.DataQualityUnavailable,
		UnsupportedReason: &reason,
		Claims:            []conversation.Claim{},
		References:        []conversation.Reference{},
	}
}

func missingScenarioAnswer() conversation.Response {
	reason := "A scenario calculation_id or simulate_scenario arguments are required."
	return conversation.Response{
		Answer:            "Run a scenario simulation first, or ask with a specific asset and change percent.",
		Intent:            conversation.IntentScenarioSimulation,
		DataQuality:       shared.DataQualityUnavailable,
		UnsupportedReason: &reason,
		Claims:            []conversation.Claim{},
		References:        []conversation.Reference{},
	}
}
