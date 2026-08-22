package coingecko

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/domain/provider"
	"github.com/maxaicrypto/backend/internal/domain/shared"
)

func TestGetPricesNormalizesQuotes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/simple/price") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ethereum":{"usd":2500.12,"usd_24h_change":-1.5}}`))
	}))
	t.Cleanup(server.Close)

	p := New(config.ProviderConfig{
		Timeout:     5 * time.Second,
		MaxAttempts: 1,
		CoinGecko:   config.ProviderCredentials{BaseURL: server.URL},
	})

	quotes, err := p.GetPrices(context.Background(), provider.PriceRequest{
		MarketDataIDs: []string{"ethereum"},
		Currency:      shared.CurrencyUSD,
	})
	if err != nil {
		t.Fatalf("GetPrices: %v", err)
	}
	if len(quotes) != 1 {
		t.Fatalf("quotes = %d, want 1", len(quotes))
	}
	if quotes[0].MarketDataID != "ethereum" {
		t.Fatalf("id = %q", quotes[0].MarketDataID)
	}
	if quotes[0].Price.String() != "2500.12" {
		t.Fatalf("price = %q", quotes[0].Price.String())
	}
	if !quotes[0].Change24hPct.Valid || quotes[0].Change24hPct.Decimal.String() != "-1.5" {
		t.Fatalf("change = %+v", quotes[0].Change24hPct)
	}
}

func TestResolveNativeMarketDataID(t *testing.T) {
	p := New(config.ProviderConfig{
		Timeout:     5 * time.Second,
		MaxAttempts: 1,
		CoinGecko:   config.ProviderCredentials{BaseURL: "http://example.invalid"},
	})
	id, ok, err := p.ResolveMarketDataID(context.Background(), provider.MappingRequest{
		ChainID:  "ethereum",
		IsNative: true,
		Symbol:   "ETH",
	})
	if err != nil {
		t.Fatalf("ResolveMarketDataID: %v", err)
	}
	if !ok || id != "ethereum" {
		t.Fatalf("got (%q, %v)", id, ok)
	}
}

func TestResolveContractMarketDataID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/coins/ethereum/contract/") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"usd-coin"}`))
	}))
	t.Cleanup(server.Close)

	p := New(config.ProviderConfig{
		Timeout:     5 * time.Second,
		MaxAttempts: 1,
		CoinGecko:   config.ProviderCredentials{BaseURL: server.URL},
	})
	contract := "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"
	id, ok, err := p.ResolveMarketDataID(context.Background(), provider.MappingRequest{
		ChainID:         "ethereum",
		ContractAddress: &contract,
		Symbol:          "USDC",
	})
	if err != nil {
		t.Fatalf("ResolveMarketDataID: %v", err)
	}
	if !ok || id != "usd-coin" {
		t.Fatalf("got (%q, %v)", id, ok)
	}
}
