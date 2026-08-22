package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	appportfolio "github.com/maxaicrypto/backend/internal/application/portfolio"
	apppricing "github.com/maxaicrypto/backend/internal/application/pricing"
	appscenarios "github.com/maxaicrypto/backend/internal/application/scenarios"
	apptransactions "github.com/maxaicrypto/backend/internal/application/transactions"
	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/conversation"
	"github.com/maxaicrypto/backend/internal/domain/performance"
	"github.com/maxaicrypto/backend/internal/domain/portfolio"
	"github.com/maxaicrypto/backend/internal/domain/provider"
	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// OrchestratorDeps wires domain services the orchestrator may invoke via tools.
type OrchestratorDeps struct {
	Portfolios   appportfolio.Service
	Performance  appportfolio.PerformanceService
	Transactions apptransactions.Service
	Pricing      apppricing.Service
	Scenarios    appscenarios.Service
	Snapshots    portfolio.SnapshotRepository
	Resolver     provider.Resolver
	AI           config.AIConfig
}

// OrchestratorApp routes intents, executes domain tools, and optionally asks the
// LLM to phrase an answer from those facts (§68). Financial numbers always come
// from tools, never from the model (§50).
type OrchestratorApp struct {
	tools      ToolRegistry
	scenarios  appscenarios.Service
	resolver   provider.Resolver
	ai         config.AIConfig
	toolBudget int
}

// NewOrchestrator wires the AI orchestrator.
func NewOrchestrator(deps OrchestratorDeps) *OrchestratorApp {
	budget := deps.AI.MaxToolCalls
	if budget < 1 {
		budget = 5
	}
	return &OrchestratorApp{
		tools: newRegistry(
			newPortfolioTool(deps.Portfolios),
			newPositionsTool(deps.Portfolios),
			newPerformanceTool(deps.Performance),
			newTransactionTool(deps.Transactions),
			newHistoricalTool(deps.Portfolios, deps.Snapshots),
			newAssetPriceTool(deps.Portfolios, deps.Pricing),
			newScenarioTool(deps.Scenarios),
		),
		scenarios:  deps.Scenarios,
		resolver:   deps.Resolver,
		ai:         deps.AI,
		toolBudget: budget,
	}
}

// Run implements Orchestrator.
func (o *OrchestratorApp) Run(ctx context.Context, req OrchestrationRequest, sink StreamSink) (RunResult, error) {
	intent := resolveIntent(req.Question, req)
	if intent == conversation.IntentUnsupported {
		response := unsupportedAnswer()
		if err := streamText(ctx, sink, response.Answer); err != nil {
			return RunResult{}, err
		}
		return RunResult{Response: response, Content: response.Answer}, nil
	}

	invocation := ToolInvocation{
		UserID:   req.UserID,
		WalletID: req.WalletID,
	}

	var (
		response  conversation.Response
		toolCalls []conversation.ToolCall
		factsJSON []byte
		usedTools int
	)

	run := func(name conversation.ToolName, args []byte) ([]byte, error) {
		if usedTools >= o.toolBudget {
			return nil, apperr.New(apperr.CodeAIUnavailable).WithMessage("Tool call budget exhausted.")
		}
		tool, ok := o.tools.Get(name)
		if !ok {
			return nil, apperr.New(apperr.CodeInternal)
		}
		invocation.Arguments = args
		result, call, err := runTool(ctx, tool, invocation, sink)
		toolCalls = append(toolCalls, call)
		usedTools++
		return result, err
	}

	switch intent {
	case conversation.IntentPortfolioPerformance:
		args, _ := json.Marshal(map[string]string{"period": string(performance.Period24h)})
		result, err := run(conversation.ToolGetPortfolioPerformance, args)
		if err != nil {
			return RunResult{}, err
		}
		factsJSON = result
		var facts map[string]any
		if err := json.Unmarshal(result, &facts); err != nil {
			return RunResult{}, err
		}
		response = portfolioPerformanceAnswer(facts, readDataQuality(facts))

	case conversation.IntentTransactionExplain:
		if req.TransactionID == nil {
			response = missingTransactionAnswer()
			break
		}
		args, _ := json.Marshal(map[string]string{"transaction_id": req.TransactionID.String()})
		result, err := run(conversation.ToolGetTransaction, args)
		if err != nil {
			return RunResult{}, err
		}
		factsJSON = result
		var facts map[string]any
		if err := json.Unmarshal(result, &facts); err != nil {
			return RunResult{}, err
		}
		response = transactionExplainAnswer(facts, readDataQuality(facts))

	case conversation.IntentScenarioSimulation:
		if req.ScenarioID != nil {
			result, err := o.scenarios.Get(ctx, req.UserID, *req.ScenarioID)
			if err != nil {
				return RunResult{}, err
			}
			encoded, err := json.Marshal(mapScenarioFacts(result))
			if err != nil {
				return RunResult{}, err
			}
			factsJSON = encoded
			var facts map[string]any
			if err := json.Unmarshal(encoded, &facts); err != nil {
				return RunResult{}, err
			}
			response = scenarioExplainAnswer(facts, readDataQuality(facts))
			break
		}
		response = missingScenarioAnswer()

	case conversation.IntentPortfolioSummary, conversation.IntentPortfolioAllocation, conversation.IntentGeneralQuestion:
		toolName := conversation.ToolGetPortfolio
		if containsAny(strings.ToLower(req.Question), "history", "historical", "snapshot") {
			toolName = conversation.ToolGetHistoricalPortfolio
		} else if containsAny(strings.ToLower(req.Question), "allocation", "largest", "concentration", "position") {
			toolName = conversation.ToolGetPositions
		}
		result, err := run(toolName, []byte(`{}`))
		if err != nil {
			return RunResult{}, err
		}
		factsJSON = result
		var facts map[string]any
		if err := json.Unmarshal(result, &facts); err != nil {
			return RunResult{}, err
		}
		response = portfolioSummaryAnswer(facts, readDataQuality(facts))
		response.Intent = intent

	default:
		response = unsupportedAnswer()
	}

	if answer, ok := o.explainWithLLM(ctx, req.Question, intent, response, factsJSON, sink); ok {
		response.Answer = answer
		return RunResult{
			Response:  response,
			ToolCalls: toolCalls,
			Content:   answer,
		}, nil
	}

	if response.Answer != "" {
		if err := streamText(ctx, sink, response.Answer); err != nil {
			return RunResult{}, err
		}
	}
	return RunResult{
		Response:  response,
		ToolCalls: toolCalls,
		Content:   response.Answer,
	}, nil
}

func (o *OrchestratorApp) explainWithLLM(
	ctx context.Context,
	question string,
	intent conversation.Intent,
	fallback conversation.Response,
	factsJSON []byte,
	sink StreamSink,
) (string, bool) {
	if !o.ai.HasAICredentials() || o.resolver == nil || len(factsJSON) == 0 {
		return "", false
	}
	llm, err := o.resolver.ResolveLLM(ctx)
	if err != nil {
		return "", false
	}

	req := provider.LLMRequest{
		Messages: []provider.LLMMessage{
			{Role: provider.RoleSystem, Content: llmSystemPrompt()},
			{Role: provider.RoleUser, Content: llmUserPrompt(question, intent, fallback.DataQuality, factsJSON)},
		},
		MaxTokens:      o.ai.MaxOutputTokens,
		ResponseSchema: llmAnswerSchema(),
	}
	// Structured answers use Generate so validation happens on complete JSON;
	// the validated answer is then streamed to the client (§74, §178).
	raw, err := llm.Generate(ctx, req)
	if err != nil {
		return "", false
	}
	answer, ok := validateLLMAnswer(raw.Content)
	if !ok {
		return "", false
	}
	if err := streamText(ctx, sink, answer); err != nil {
		return "", false
	}
	return answer, true
}

func llmSystemPrompt() string {
	return strings.Join([]string{
		"You are MaxAI Crypto, a portfolio explanation assistant.",
		"Use only the structured facts provided by the backend tools.",
		"Never invent balances, prices, percentages, or transaction details.",
		"Never give trading, investment, or leverage advice.",
		"If data_quality is not COMPLETE, say so briefly.",
		"Respond with JSON matching the schema: {\"answer\":\"...\"}.",
	}, " ")
}

func llmUserPrompt(question string, intent conversation.Intent, quality shared.DataQuality, factsJSON []byte) string {
	return fmt.Sprintf(
		"Intent: %s\nData quality: %s\nQuestion: %s\n\nFacts JSON:\n%s",
		intent,
		quality,
		question,
		string(factsJSON),
	)
}

func llmAnswerSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"answer": map[string]any{"type": "string"},
		},
		"required":             []string{"answer"},
		"additionalProperties": false,
	}
}

func validateLLMAnswer(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}
	if strings.HasPrefix(trimmed, "{") {
		var payload struct {
			Answer string `json:"answer"`
		}
		if err := json.Unmarshal([]byte(trimmed), &payload); err == nil {
			answer := strings.TrimSpace(payload.Answer)
			if answer != "" {
				return answer, true
			}
		}
		return "", false
	}
	return trimmed, true
}

func readDataQuality(facts map[string]any) shared.DataQuality {
	if raw, ok := facts["data_quality"].(string); ok && raw != "" {
		return shared.DataQuality(raw)
	}
	return shared.DataQualityComplete
}

var _ Orchestrator = (*OrchestratorApp)(nil)
