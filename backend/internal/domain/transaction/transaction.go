// Package transaction models canonical blockchain transactions. Raw provider
// responses are never the domain source of truth (§45).
package transaction

import (
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// Type is the backend-determined transaction classification (§46). When the
// classifier is not confident the value stays TypeUnknown; the LLM may never
// promote it to a confirmed type without backend evidence (§47).
type Type string

const (
	TypeTransfer            Type = "TRANSFER"
	TypeSwap                Type = "SWAP"
	TypeStake               Type = "STAKE"
	TypeUnstake             Type = "UNSTAKE"
	TypeClaim               Type = "CLAIM"
	TypeApprove             Type = "APPROVE"
	TypeContractInteraction Type = "CONTRACT_INTERACTION"
	TypeUnknown             Type = "UNKNOWN"
)

// Status is the on-chain execution result.
type Status string

const (
	StatusSuccess Status = "SUCCESS"
	StatusFailed  Status = "FAILED"
	StatusPending Status = "PENDING"
)

// Transaction is the canonical transaction record (§45). Only normalized facts
// and a provider reference are stored, never a full provider payload (§163).
type Transaction struct {
	ID       uuid.UUID
	WalletID uuid.UUID
	ChainID  chain.ID

	TxHash      string
	BlockNumber *int64
	Timestamp   time.Time

	Status Status
	Type   Type

	FromAddress *string
	ToAddress   *string

	AssetInID *uuid.UUID
	AmountIn  shared.NullDecimal

	AssetOutID *uuid.UUID
	AmountOut  shared.NullDecimal

	FeeAssetID *uuid.UUID
	FeeAmount  shared.NullDecimal

	Protocol     *string
	Counterparty *string

	// RawReference points back to the provider record this transaction was
	// derived from, for debugging and re-fetching. It is a reference, not a
	// stored payload (§163).
	RawReference *string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Identity is the uniqueness key that makes a retried sync idempotent (§48).
// A log or transfer index is included where provider semantics require it,
// because one hash can produce several wallet-relevant movements.
type Identity struct {
	WalletID uuid.UUID
	ChainID  chain.ID
	TxHash   string
	// LogIndex disambiguates several movements inside one transaction. It is
	// zero for chains and providers where a hash maps to a single movement.
	LogIndex int
}

// Filter narrows a transaction listing.
type Filter struct {
	// Type restricts results to one classification when set.
	Type *Type
}
