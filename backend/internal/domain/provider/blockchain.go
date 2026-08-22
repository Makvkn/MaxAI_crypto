package provider

import (
	"context"
	"time"

	"github.com/maxaicrypto/backend/internal/domain/chain"
)

// BalanceRequest asks a provider for a wallet's current holdings.
type BalanceRequest struct {
	ChainID chain.ID
	// Address is already normalized to the chain's canonical form (§188).
	Address string
}

// TransactionRequest asks a provider for a wallet's transaction history.
type TransactionRequest struct {
	ChainID chain.ID
	Address string
	// Limit caps the number of records requested in one call.
	Limit int
	// PageToken is the provider's own pagination token from a previous
	// response. It is opaque to the domain and never reaches the API.
	PageToken *string
	// Since restricts results to transactions at or after this instant, which
	// keeps incremental syncs cheap.
	Since *time.Time
}

// TransactionPage is one page of normalized transactions.
type TransactionPage struct {
	Transactions []NormalizedTransaction
	// NextPageToken is nil when the provider has no further pages.
	NextPageToken *string
}

// BlockchainDataProvider is the port every blockchain adapter implements
// (§22). Implementations translate their own DTOs into the normalized types in
// this package and map their errors into apperr codes before returning (§28).
type BlockchainDataProvider interface {
	// Name identifies the implementation for logging, metrics and routing.
	Name() Name
	// Capabilities reports which chains and operations this provider serves.
	Capabilities(ctx context.Context) Capabilities
	// GetBalances returns the wallet's current holdings.
	GetBalances(ctx context.Context, req BalanceRequest) ([]NormalizedBalance, error)
	// GetTransactions returns one page of the wallet's transaction history.
	GetTransactions(ctx context.Context, req TransactionRequest) (TransactionPage, error)
}
