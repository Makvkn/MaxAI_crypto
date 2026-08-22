// Package address validates and normalizes blockchain addresses per chain.
package address

import (
	"regexp"
	"strings"

	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/chain"
)

// Validator implements chain.AddressValidator.
type Validator struct{}

// NewValidator builds an address validator.
func NewValidator() *Validator {
	return &Validator{}
}

var (
	evmRe       = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
	solanaRe    = regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]{32,44}$`)
	tronRe      = regexp.MustCompile(`^T[1-9A-HJ-NP-Za-km-z]{33}$`)
	xrplRe      = regexp.MustCompile(`^r[1-9A-HJ-NP-Za-km-z]{24,34}$`)
	btcLegacyRe = regexp.MustCompile(`^[13][a-km-zA-HJ-NP-Z1-9]{25,34}$`)
	bech32Re    = regexp.MustCompile(`^(bc1|ltc1|doge1)[a-z0-9]{25,90}$`)
)

// Normalize implements chain.AddressValidator.
func (v *Validator) Normalize(id chain.ID, address string) (string, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return "", apperr.New(apperr.CodeInvalidWalletAddress)
	}

	format := formatFor(id)
	switch format {
	case chain.AddressFormatEVM:
		if !evmRe.MatchString(address) {
			return "", apperr.New(apperr.CodeInvalidWalletAddress)
		}
		return strings.ToLower(address), nil
	case chain.AddressFormatSolana:
		if !solanaRe.MatchString(address) {
			return "", apperr.New(apperr.CodeInvalidWalletAddress)
		}
		return address, nil
	case chain.AddressFormatTron:
		if !tronRe.MatchString(address) {
			return "", apperr.New(apperr.CodeInvalidWalletAddress)
		}
		return address, nil
	case chain.AddressFormatXRPL:
		if !xrplRe.MatchString(address) {
			return "", apperr.New(apperr.CodeInvalidWalletAddress)
		}
		return address, nil
	case chain.AddressFormatBitcoinLike:
		lower := strings.ToLower(address)
		if btcLegacyRe.MatchString(address) || bech32Re.MatchString(lower) {
			if strings.HasPrefix(lower, "bc1") || strings.HasPrefix(lower, "ltc1") || strings.HasPrefix(lower, "doge1") {
				return lower, nil
			}
			return address, nil
		}
		return "", apperr.New(apperr.CodeInvalidWalletAddress)
	default:
		return "", apperr.New(apperr.CodeUnsupportedChain)
	}
}

func formatFor(id chain.ID) chain.AddressFormat {
	switch id {
	case chain.Ethereum, chain.BNBChain:
		return chain.AddressFormatEVM
	case chain.Bitcoin, chain.Litecoin, chain.Dogecoin:
		return chain.AddressFormatBitcoinLike
	case chain.Solana:
		return chain.AddressFormatSolana
	case chain.Tron:
		return chain.AddressFormatTron
	case chain.XRPLedger:
		return chain.AddressFormatXRPL
	default:
		return ""
	}
}

var _ chain.AddressValidator = (*Validator)(nil)
