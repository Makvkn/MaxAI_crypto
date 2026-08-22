package transaction

import (
	"context"
	"strings"
)

// RulesClassifier assigns transaction types from deterministic evidence (§47).
// Uncertainty stays TypeUnknown; the LLM never promotes a type without this.
type RulesClassifier struct{}

// NewRulesClassifier builds the default classifier.
func NewRulesClassifier() *RulesClassifier { return &RulesClassifier{} }

// Classify implements Classifier.
func (c *RulesClassifier) Classify(_ context.Context, tx Transaction) (Type, error) {
	hasIn := tx.AssetInID != nil && tx.AmountIn.Valid
	hasOut := tx.AssetOutID != nil && tx.AmountOut.Valid

	if hasIn && hasOut {
		return TypeSwap, nil
	}

	if hint := protocolHint(tx.Protocol); hint != TypeUnknown {
		return hint, nil
	}

	if hasIn || hasOut {
		return TypeTransfer, nil
	}

	if tx.ToAddress != nil && *tx.ToAddress != "" {
		return TypeContractInteraction, nil
	}

	return TypeUnknown, nil
}

func protocolHint(protocol *string) Type {
	if protocol == nil {
		return TypeUnknown
	}
	value := strings.ToLower(strings.TrimSpace(*protocol))
	switch {
	case strings.Contains(value, "unstake"):
		return TypeUnstake
	case strings.Contains(value, "stake"):
		return TypeStake
	case strings.Contains(value, "claim") || strings.Contains(value, "reward"):
		return TypeClaim
	case strings.Contains(value, "approve") || strings.Contains(value, "permit"):
		return TypeApprove
	case strings.Contains(value, "swap") || strings.Contains(value, "trade") || strings.Contains(value, "exchange"):
		return TypeSwap
	case strings.Contains(value, "transfer"):
		return TypeTransfer
	default:
		return TypeUnknown
	}
}

var _ Classifier = (*RulesClassifier)(nil)
