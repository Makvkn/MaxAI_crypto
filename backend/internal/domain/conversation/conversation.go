// Package conversation models AI conversations and the structured AI response
// contract. Every conversation belongs to a user and, in the MVP, to the wallet
// being analysed (§181, §182).
package conversation

import (
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// Conversation is a dialogue about one wallet's portfolio.
type Conversation struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	WalletID uuid.UUID
	Title    string
	// MessageCount and LastMessagePreview are denormalized for list rendering
	// so the listing does not load every message.
	MessageCount       int
	LastMessagePreview *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Role identifies the author of a message.
type Role string

const (
	RoleUser      Role = "USER"
	RoleAssistant Role = "ASSISTANT"
)

// MessageStatus tracks an assistant message through its streamed lifecycle
// (§175).
type MessageStatus string

const (
	MessagePending   MessageStatus = "PENDING"
	MessageStreaming MessageStatus = "STREAMING"
	MessageCompleted MessageStatus = "COMPLETED"
	MessageFailed    MessageStatus = "FAILED"
)

// Message is a single conversation turn.
type Message struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	Role           Role
	Status         MessageStatus
	Content        string
	// Response holds the structured assistant result. It is nil for user
	// messages and for assistant messages that never completed.
	Response  *Response
	ToolCalls []ToolCall
	// ErrorCode records a domain-level failure so the frontend can render the
	// standard error envelope inside the message.
	ErrorCode    *string
	ErrorMessage *string
	CreatedAt    time.Time
}

// Intent is the routing decision for a user question (§69). Routing uses a
// registry rather than a large if/else tree.
type Intent string

const (
	IntentPortfolioSummary     Intent = "PORTFOLIO_SUMMARY"
	IntentPortfolioPerformance Intent = "PORTFOLIO_PERFORMANCE"
	IntentPortfolioAllocation  Intent = "PORTFOLIO_ALLOCATION"
	IntentTransactionExplain   Intent = "TRANSACTION_EXPLANATION"
	IntentScenarioSimulation   Intent = "SCENARIO_SIMULATION"
	IntentGeneralQuestion      Intent = "GENERAL_PORTFOLIO_QUESTION"
	// IntentUnsupported covers questions outside MVP scope, such as news
	// intelligence. The answer says so instead of hallucinating (§76).
	IntentUnsupported Intent = "UNSUPPORTED"
)

// EvidenceType names the kind of backend fact a claim rests on (§73).
type EvidenceType string

const (
	EvidenceCalculation         EvidenceType = "calculation"
	EvidencePortfolio           EvidenceType = "portfolio"
	EvidencePortfolioPerformace EvidenceType = "portfolio_performance"
	EvidenceSnapshot            EvidenceType = "portfolio_snapshot"
	EvidencePosition            EvidenceType = "position"
	EvidenceTransaction         EvidenceType = "transaction"
	EvidencePrice               EvidenceType = "price"
	EvidenceScenario            EvidenceType = "scenario"
)

// Evidence links a claim to the backend computation that supports it.
type Evidence struct {
	Type EvidenceType
	ID   string
}

// Claim is one factual assertion together with its supporting evidence. The AI
// may reason freely but must not present unsupported assumptions as verified
// facts (§73).
type Claim struct {
	Text     string
	Evidence []Evidence
}

// ReferenceType names an entity a response points at (§180).
type ReferenceType string

const (
	ReferenceAsset       ReferenceType = "asset"
	ReferenceTransaction ReferenceType = "transaction"
	ReferencePortfolio   ReferenceType = "portfolio"
	ReferenceScenario    ReferenceType = "scenario"
)

// Reference points at an entity the frontend can link to.
type Reference struct {
	Type  ReferenceType
	ID    string
	Label *string
}

// Response is the structured AI result (§74). It is an object, never a bare
// string, so the frontend can build richer evidence-aware UI later.
type Response struct {
	Answer string
	Intent Intent
	// DataQuality must reflect the true quality of the facts used. The AI may
	// not describe partial or stale data as fully current (§143, §144).
	DataQuality shared.DataQuality
	Claims      []Claim
	References  []Reference
	// UnsupportedReason explains an UNSUPPORTED intent.
	UnsupportedReason *string
}

// ToolName is a domain tool the orchestrator may execute on the LLM's behalf
// (§70).
type ToolName string

const (
	ToolGetPortfolio            ToolName = "get_portfolio"
	ToolGetPositions            ToolName = "get_positions"
	ToolGetPortfolioPerformance ToolName = "get_portfolio_performance"
	ToolGetTransaction          ToolName = "get_transaction"
	ToolGetHistoricalPortfolio  ToolName = "get_historical_portfolio"
	ToolGetAssetPrice           ToolName = "get_asset_price"
	ToolSimulateScenario        ToolName = "simulate_scenario"
)

// ToolCallStatus tracks one tool execution.
type ToolCallStatus string

const (
	ToolCallRunning   ToolCallStatus = "RUNNING"
	ToolCallCompleted ToolCallStatus = "COMPLETED"
	ToolCallFailed    ToolCallStatus = "FAILED"
)

// ToolCall records one tool execution inside a message, which the SSE stream
// surfaces as tool_started and tool_completed events (§81).
type ToolCall struct {
	ID          string
	Tool        ToolName
	Status      ToolCallStatus
	StartedAt   time.Time
	CompletedAt *time.Time
}
