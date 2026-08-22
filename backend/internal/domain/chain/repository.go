package chain

import "context"

// Repository exposes the chain lookups the application needs. It is not a
// generic CRUD interface (§115).
type Repository interface {
	// List returns every chain known to the system, including unsupported ones.
	List(ctx context.Context) ([]Chain, error)
	// GetByID returns one chain, or apperr.CodeUnsupportedChain when the
	// identifier is not a chain the backend knows.
	GetByID(ctx context.Context, id ID) (Chain, error)
}

// AddressValidator normalizes and validates an address for a specific chain.
// Validation is backend-authoritative; the frontend only provides immediate UX
// feedback (§188, §189).
type AddressValidator interface {
	// Normalize returns the canonical representation of address on the given
	// chain, so that two spellings of the same logical address cannot create
	// duplicate wallets.
	Normalize(id ID, address string) (string, error)
}
