// Package chain models supported blockchain networks. Chains are domain
// entities backed by configuration, never hardcoded branches inside portfolio
// algorithms (§20).
package chain

import (
	"time"

	"github.com/google/uuid"
)

// ID is the stable machine identifier of a chain. Display names are never used
// as identifiers (§21).
//
// The codes below are the wire values the frontend already sends. The
// specification's §21 examples spell two of them differently (bnb_chain,
// xrp_ledger); the divergence and its resolution are recorded in
// openapi/DECISIONS.md.
type ID string

const (
	Ethereum  ID = "ethereum"
	Bitcoin   ID = "bitcoin"
	BNBChain  ID = "bnb"
	Solana    ID = "solana"
	Litecoin  ID = "litecoin"
	XRPLedger ID = "xrpl"
	Tron      ID = "tron"
	Dogecoin  ID = "dogecoin"
)

// Supported lists the MVP chains (§20). TRON is the blockchain; TRX is its
// native asset.
var Supported = []ID{Ethereum, Bitcoin, BNBChain, Solana, Litecoin, XRPLedger, Tron, Dogecoin}

// IsSupported reports whether id is one of the MVP chains.
func IsSupported(id ID) bool {
	for _, supported := range Supported {
		if supported == id {
			return true
		}
	}
	return false
}

// Chain is a supported blockchain network.
type Chain struct {
	ID   ID
	Name string
	// NativeAssetID references the chain's native asset, which has no contract
	// address (§31).
	NativeAssetID uuid.UUID
	// AddressFormat selects the chain-specific validation and normalization
	// rules. A single generic regex across all chains is explicitly wrong (§189).
	AddressFormat AddressFormat
	IsSupported   bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// AddressFormat identifies a family of address encoding rules shared by
// several chains.
type AddressFormat string

const (
	// AddressFormatEVM covers 0x-prefixed 20-byte addresses.
	AddressFormatEVM AddressFormat = "EVM"
	// AddressFormatBitcoinLike covers Base58Check and Bech32 UTXO addresses.
	AddressFormatBitcoinLike AddressFormat = "BITCOIN_LIKE"
	// AddressFormatSolana covers Base58 Ed25519 public keys.
	AddressFormatSolana AddressFormat = "SOLANA"
	// AddressFormatTron covers Base58Check addresses starting with T.
	AddressFormatTron AddressFormat = "TRON"
	// AddressFormatXRPL covers classic XRP Ledger account addresses.
	AddressFormatXRPL AddressFormat = "XRPL"
)
