package apperr

import "net/http"

// Code is a stable, frontend-facing error identifier. Provider-specific
// details never appear here (§28, §105).
type Code string

// Category groups codes for transport-level handling (§105).
type Category string

const (
	CategoryValidation     Category = "VALIDATION_ERROR"
	CategoryAuthentication Category = "AUTHENTICATION_ERROR"
	CategoryNotFound       Category = "NOT_FOUND"
	CategoryProvider       Category = "PROVIDER_ERROR"
	CategoryDataUnavail    Category = "DATA_UNAVAILABLE"
	CategoryRateLimit      Category = "RATE_LIMIT"
	CategoryInternal       Category = "INTERNAL_ERROR"
)

// Domain error codes (§106). Spellings match the contract the frontend already
// consumes; codes the frontend does not yet handle are additive and carry a
// category that the frontend falls back on.
const (
	// Validation.
	CodeValidation           Code = "VALIDATION_ERROR"
	CodeInvalidWalletAddress Code = "INVALID_WALLET_ADDRESS"
	CodeUnsupportedChain     Code = "UNSUPPORTED_CHAIN"

	// Authentication.
	CodeAuthentication         Code = "AUTHENTICATION_ERROR"
	CodeSessionExpired         Code = "SESSION_EXPIRED"
	CodeEmailAlreadyRegistered Code = "EMAIL_ALREADY_REGISTERED"
	CodeInvalidCredentials     Code = "INVALID_CREDENTIALS"
	CodeForbidden              Code = "FORBIDDEN"

	// Not found.
	CodeNotFound             Code = "NOT_FOUND"
	CodeWalletNotFound       Code = "WALLET_NOT_FOUND"
	CodeTransactionNotFound  Code = "TRANSACTION_NOT_FOUND"
	CodeConversationNotFound Code = "CONVERSATION_NOT_FOUND"
	CodeAssetNotFound        Code = "ASSET_NOT_FOUND"

	// Provider.
	CodeProviderError Code = "PROVIDER_ERROR"

	// Data availability. STALE describes freshness, PARTIAL describes
	// completeness, UNAVAILABLE describes missing required data (§42).
	CodeDataUnavailable                     Code = "DATA_UNAVAILABLE"
	CodePortfolioDataUnavailable            Code = "PORTFOLIO_DATA_UNAVAILABLE"
	CodePortfolioDataTemporarilyUnavailable Code = "PORTFOLIO_DATA_TEMPORARILY_UNAVAILABLE"
	CodePortfolioDataPartial                Code = "PORTFOLIO_DATA_PARTIAL"
	CodePortfolioDataStale                  Code = "PORTFOLIO_DATA_STALE"
	CodePortfolioValuationUnavailable       Code = "PORTFOLIO_VALUATION_UNAVAILABLE"
	CodePerformanceDataUnavailable          Code = "PERFORMANCE_DATA_UNAVAILABLE"
	CodePriceDataUnavailable                Code = "PRICE_DATA_UNAVAILABLE"

	// Wallet lifecycle and synchronization.
	CodeWalletNotReady       Code = "WALLET_NOT_READY"
	CodeWalletSyncFailed     Code = "WALLET_SYNC_FAILED"
	CodeWalletSyncInProgress Code = "WALLET_SYNC_IN_PROGRESS"

	// Transactions.
	CodeTransactionClassificationUnknown Code = "TRANSACTION_CLASSIFICATION_UNKNOWN"

	// Rate limits and entitlements.
	CodeRateLimit          Code = "RATE_LIMIT"
	CodeAIDailyLimit       Code = "AI_DAILY_LIMIT_REACHED"
	CodeWalletLimitReached Code = "WALLET_LIMIT_REACHED"

	// AI.
	CodeAIUnavailable        Code = "AI_UNAVAILABLE"
	CodeAIStreamFailed       Code = "AI_STREAM_FAILED"
	CodeAIUnsupportedIntent  Code = "AI_UNSUPPORTED_INTENT"
	CodeAIContextUnavailable Code = "AI_CONTEXT_UNAVAILABLE"
	CodeAIInvalidResponse    Code = "AI_PROVIDER_INVALID_RESPONSE"

	// Scenarios.
	CodeScenarioInvalid          Code = "SCENARIO_INVALID"
	CodeScenarioAssetUnavailable Code = "SCENARIO_ASSET_UNAVAILABLE"

	// Internal.
	CodeInternal       Code = "INTERNAL_ERROR"
	CodeNotImplemented Code = "NOT_IMPLEMENTED"
)

// codeMeta describes how a code is categorised and surfaced over HTTP.
type codeMeta struct {
	category Category
	status   int
	message  string
}

// registry is the single place where a code's category, HTTP status and
// default user-facing message are defined (§106: keep the list central).
var registry = map[Code]codeMeta{
	CodeValidation:           {CategoryValidation, http.StatusBadRequest, "The request is invalid."},
	CodeInvalidWalletAddress: {CategoryValidation, http.StatusBadRequest, "The wallet address is not valid for the selected chain."},
	CodeUnsupportedChain:     {CategoryValidation, http.StatusBadRequest, "This blockchain is not supported."},

	CodeAuthentication:         {CategoryAuthentication, http.StatusUnauthorized, "Authentication is required."},
	CodeSessionExpired:         {CategoryAuthentication, http.StatusUnauthorized, "The session has expired."},
	CodeEmailAlreadyRegistered: {CategoryValidation, http.StatusConflict, "This email is already registered."},
	CodeInvalidCredentials:     {CategoryAuthentication, http.StatusUnauthorized, "The email or password is incorrect."},
	CodeForbidden:              {CategoryAuthentication, http.StatusForbidden, "Access to this resource is not allowed."},

	CodeNotFound:             {CategoryNotFound, http.StatusNotFound, "The requested resource was not found."},
	CodeWalletNotFound:       {CategoryNotFound, http.StatusNotFound, "The wallet was not found."},
	CodeTransactionNotFound:  {CategoryNotFound, http.StatusNotFound, "The transaction was not found."},
	CodeConversationNotFound: {CategoryNotFound, http.StatusNotFound, "The conversation was not found."},
	CodeAssetNotFound:        {CategoryNotFound, http.StatusNotFound, "The asset was not found."},

	CodeProviderError: {CategoryProvider, http.StatusBadGateway, "An upstream data source is currently unavailable."},

	CodeDataUnavailable:                     {CategoryDataUnavail, http.StatusServiceUnavailable, "The requested data is currently unavailable."},
	CodePortfolioDataUnavailable:            {CategoryDataUnavail, http.StatusServiceUnavailable, "Portfolio data is unavailable."},
	CodePortfolioDataTemporarilyUnavailable: {CategoryDataUnavail, http.StatusServiceUnavailable, "Portfolio data is temporarily unavailable."},
	CodePortfolioDataPartial:                {CategoryDataUnavail, http.StatusServiceUnavailable, "Portfolio data is incomplete."},
	CodePortfolioDataStale:                  {CategoryDataUnavail, http.StatusServiceUnavailable, "Portfolio data is stale."},
	CodePortfolioValuationUnavailable:       {CategoryDataUnavail, http.StatusServiceUnavailable, "Portfolio valuation could not be produced."},
	CodePerformanceDataUnavailable:          {CategoryDataUnavail, http.StatusServiceUnavailable, "Performance data is unavailable for this period."},
	CodePriceDataUnavailable:                {CategoryDataUnavail, http.StatusServiceUnavailable, "Price data is unavailable."},

	CodeWalletNotReady:       {CategoryDataUnavail, http.StatusConflict, "The wallet is still synchronizing."},
	CodeWalletSyncFailed:     {CategoryDataUnavail, http.StatusServiceUnavailable, "Wallet synchronization failed."},
	CodeWalletSyncInProgress: {CategoryDataUnavail, http.StatusConflict, "A synchronization is already running for this wallet."},

	CodeTransactionClassificationUnknown: {CategoryDataUnavail, http.StatusOK, "The transaction type could not be determined."},

	CodeRateLimit:          {CategoryRateLimit, http.StatusTooManyRequests, "Too many requests. Please try again later."},
	CodeAIDailyLimit:       {CategoryRateLimit, http.StatusTooManyRequests, "The daily AI operation limit has been reached."},
	CodeWalletLimitReached: {CategoryRateLimit, http.StatusForbidden, "The wallet limit for this plan has been reached."},

	CodeAIUnavailable:        {CategoryProvider, http.StatusServiceUnavailable, "AI analysis is currently unavailable."},
	CodeAIStreamFailed:       {CategoryProvider, http.StatusServiceUnavailable, "The AI response stream was interrupted."},
	CodeAIUnsupportedIntent:  {CategoryValidation, http.StatusUnprocessableEntity, "This question is not supported yet."},
	CodeAIContextUnavailable: {CategoryDataUnavail, http.StatusServiceUnavailable, "The data required to answer is unavailable."},
	CodeAIInvalidResponse:    {CategoryProvider, http.StatusBadGateway, "The AI response could not be validated."},

	CodeScenarioInvalid:          {CategoryValidation, http.StatusBadRequest, "The scenario parameters are invalid."},
	CodeScenarioAssetUnavailable: {CategoryDataUnavail, http.StatusServiceUnavailable, "The scenario cannot be calculated for this asset."},

	CodeInternal:       {CategoryInternal, http.StatusInternalServerError, "An unexpected error occurred."},
	CodeNotImplemented: {CategoryInternal, http.StatusNotImplemented, "This operation is not implemented yet."},
}

func metaFor(code Code) codeMeta {
	if meta, ok := registry[code]; ok {
		return meta
	}
	return registry[CodeInternal]
}

// CategoryOf reports the category of a code.
func CategoryOf(code Code) Category { return metaFor(code).category }

// StatusOf reports the HTTP status a code maps to.
func StatusOf(code Code) int { return metaFor(code).status }

// IsKnown reports whether the code is registered.
func IsKnown(code Code) bool {
	_, ok := registry[code]
	return ok
}
