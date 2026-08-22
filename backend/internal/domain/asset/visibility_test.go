package asset

import (
	"context"
	"testing"
)

func TestRulesVisibilitySpamByName(t *testing.T) {
	c := NewRulesVisibilityClassifier()
	got, err := c.Classify(context.Background(), VisibilityInput{
		Asset:         Asset{Symbol: "XYZ", Name: "Visit https://scam.example for reward", Type: TypeToken},
		HasMarketData: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != VisibilityHiddenSpam {
		t.Fatalf("got %s, want HIDDEN_SPAM", got)
	}
}

func TestRulesVisibilityDust(t *testing.T) {
	c := NewRulesVisibilityClassifier()
	value := "0.001"
	got, err := c.Classify(context.Background(), VisibilityInput{
		Asset:         Asset{Symbol: "ETH", Name: "Ether", Type: TypeNative},
		ValueUSD:      &value,
		HasMarketData: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != VisibilityHiddenDust {
		t.Fatalf("got %s, want HIDDEN_DUST", got)
	}
}

func TestRulesVisibilityVisible(t *testing.T) {
	c := NewRulesVisibilityClassifier()
	value := "12.5"
	got, err := c.Classify(context.Background(), VisibilityInput{
		Asset:         Asset{Symbol: "BTC", Name: "Bitcoin", Type: TypeNative},
		ValueUSD:      &value,
		HasMarketData: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != VisibilityVisible {
		t.Fatalf("got %s, want VISIBLE", got)
	}
}
