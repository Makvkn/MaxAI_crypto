package tatum

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/provider"
)

func TestGetBalancesUTXO(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/bitcoin/address/balance/") {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Fatalf("missing api key")
		}
		_, _ = w.Write([]byte(`{"incoming":"1.5","outgoing":"0.25"}`))
	}))
	t.Cleanup(server.Close)

	p := New(config.ProviderConfig{
		Timeout:     5 * time.Second,
		MaxAttempts: 1,
		Tatum:       config.ProviderCredentials{BaseURL: server.URL, APIKey: "test-key"},
	})
	balances, err := p.GetBalances(context.Background(), provider.BalanceRequest{
		ChainID: chain.Bitcoin,
		Address: "bc1qexample",
	})
	if err != nil {
		t.Fatalf("GetBalances: %v", err)
	}
	if len(balances) != 1 {
		t.Fatalf("len = %d", len(balances))
	}
	if balances[0].Metadata.Symbol != "BTC" {
		t.Fatalf("symbol = %s", balances[0].Metadata.Symbol)
	}
	if balances[0].BalanceNormalized.String() != "1.25" {
		t.Fatalf("balance = %s", balances[0].BalanceNormalized.String())
	}
}

func TestGetTransactionsUTXOIncoming(t *testing.T) {
	wallet := "bc1qwallet"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/v4/data/blockchains/transaction/history/utxos") {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("chain") != "bitcoin-mainnet" {
			t.Fatalf("chain = %s", r.URL.Query().Get("chain"))
		}
		if r.URL.Query().Get("address") != wallet {
			t.Fatalf("address = %s", r.URL.Query().Get("address"))
		}
		_, _ = w.Write([]byte(`[
		  {
		    "hash":"abc123",
		    "blockNumber":100,
		    "time":1700000000,
		    "fee":726,
		    "inputs":[{"address":"bc1qsender","coin":{"value":200000000,"address":"bc1qsender"}}],
		    "outputs":[{"value":150000000,"address":"bc1qwallet"}]
		  }
		]`))
	}))
	t.Cleanup(server.Close)

	p := New(config.ProviderConfig{
		Timeout:     5 * time.Second,
		MaxAttempts: 1,
		Tatum:       config.ProviderCredentials{BaseURL: server.URL, APIKey: "test-key"},
	})
	page, err := p.GetTransactions(context.Background(), provider.TransactionRequest{
		ChainID: chain.Bitcoin,
		Address: wallet,
		Limit:   20,
	})
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}
	if len(page.Transactions) != 1 {
		t.Fatalf("len = %d", len(page.Transactions))
	}
	tx := page.Transactions[0]
	if tx.TxHash != "abc123" {
		t.Fatalf("hash = %s", tx.TxHash)
	}
	if len(tx.Transfers) != 1 || tx.Transfers[0].Direction != provider.DirectionIn {
		t.Fatalf("transfers = %+v", tx.Transfers)
	}
	if tx.Transfers[0].Amount.String() != "1.5" {
		t.Fatalf("amount = %s", tx.Transfers[0].Amount.String())
	}
	if !tx.FeeAmount.Valid || tx.FeeAmount.Decimal.String() != "0.00000726" {
		t.Fatalf("fee = %+v", tx.FeeAmount)
	}
}
