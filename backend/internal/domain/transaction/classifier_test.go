package transaction

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/shared"
)

func TestRulesClassifierTransfer(t *testing.T) {
	c := NewRulesClassifier()
	assetID := uuid.New()
	got, err := c.Classify(context.Background(), Transaction{
		AssetInID: &assetID,
		AmountIn:  shared.Known(shared.NewDecimalFromInt(1)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != TypeTransfer {
		t.Fatalf("got %s, want TRANSFER", got)
	}
}

func TestRulesClassifierSwapWhenBothLegs(t *testing.T) {
	c := NewRulesClassifier()
	in := uuid.New()
	out := uuid.New()
	got, err := c.Classify(context.Background(), Transaction{
		AssetInID:  &in,
		AmountIn:   shared.Known(shared.NewDecimalFromInt(1)),
		AssetOutID: &out,
		AmountOut:  shared.Known(shared.NewDecimalFromInt(2)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != TypeSwap {
		t.Fatalf("got %s, want SWAP", got)
	}
}

func TestRulesClassifierProtocolStake(t *testing.T) {
	c := NewRulesClassifier()
	protocol := "lido-stake"
	got, err := c.Classify(context.Background(), Transaction{Protocol: &protocol})
	if err != nil {
		t.Fatal(err)
	}
	if got != TypeStake {
		t.Fatalf("got %s, want STAKE", got)
	}
}

func TestRulesClassifierUnknownWithoutEvidence(t *testing.T) {
	c := NewRulesClassifier()
	got, err := c.Classify(context.Background(), Transaction{})
	if err != nil {
		t.Fatal(err)
	}
	if got != TypeUnknown {
		t.Fatalf("got %s, want UNKNOWN", got)
	}
}
