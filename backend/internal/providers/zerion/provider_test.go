package zerion

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

func TestGetBalancesNormalizesNativeAndToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/positions/") {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); !strings.HasPrefix(got, "Basic ") {
			t.Fatalf("missing basic auth: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
		  "data": [
		    {
		      "id": "eth-pos",
		      "attributes": {
		        "quantity": {"int":"1500000000000000000","decimals":18,"numeric":"1.5"},
		        "fungible_info": {
		          "name":"Ether","symbol":"ETH","icon":null,
		          "implementations":[{"chain_id":"ethereum","address":null,"decimals":18}]
		        },
		        "updated_at":"2024-01-02T03:04:05Z",
		        "position_type":"wallet"
		      }
		    },
		    {
		      "id":"usdc-pos",
		      "attributes": {
		        "quantity":{"int":"2500000","decimals":6,"numeric":"2.5"},
		        "fungible_info":{
		          "name":"USD Coin","symbol":"USDC","icon":{"url":"https://example/usdc.png"},
		          "implementations":[{"chain_id":"ethereum","address":"0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48","decimals":6}]
		        },
		        "updated_at":"2024-01-02T03:04:05Z",
		        "position_type":"wallet"
		      }
		    }
		  ]
		}`))
	}))
	t.Cleanup(server.Close)

	p := New(config.ProviderConfig{
		Timeout:     5 * time.Second,
		MaxAttempts: 1,
		Zerion:      config.ProviderCredentials{BaseURL: server.URL, APIKey: "test-key"},
	})

	balances, err := p.GetBalances(context.Background(), provider.BalanceRequest{
		ChainID: chain.Ethereum,
		Address: "0xabc",
	})
	if err != nil {
		t.Fatalf("GetBalances: %v", err)
	}
	if len(balances) != 2 {
		t.Fatalf("len = %d", len(balances))
	}
	if balances[0].Metadata.Symbol != "ETH" || balances[0].AssetIdentity.ContractAddress != nil {
		t.Fatalf("native = %+v", balances[0])
	}
	if balances[1].Metadata.Symbol != "USDC" || balances[1].AssetIdentity.ContractAddress == nil {
		t.Fatalf("token = %+v", balances[1])
	}
	if balances[1].BalanceNormalized.String() != "2.5" {
		t.Fatalf("balance = %s", balances[1].BalanceNormalized.String())
	}
}

func TestGetTransactionsNormalizesTransfer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
		  "links": {"next":"https://api.zerion.io/v1/wallets/0x/transactions/?page[after]=cursor1"},
		  "data": [{
		    "id":"tx1",
		    "attributes":{
		      "hash":"0xhash",
		      "mined_at_block":123,
		      "mined_at":"2024-01-02T03:04:05Z",
		      "sent_from":"0xfrom",
		      "sent_to":"0xto",
		      "status":"confirmed",
		      "transfers":[{
		        "direction":"in",
		        "quantity":{"int":"100000000000000000","decimals":18,"numeric":"0.1"},
		        "fungible_info":{
		          "name":"Ether","symbol":"ETH",
		          "implementations":[{"chain_id":"ethereum","address":null,"decimals":18}]
		        }
		      }]
		    }
		  }]
		}`))
	}))
	t.Cleanup(server.Close)

	p := New(config.ProviderConfig{
		Timeout:     5 * time.Second,
		MaxAttempts: 1,
		Zerion:      config.ProviderCredentials{BaseURL: server.URL, APIKey: "test-key"},
	})
	page, err := p.GetTransactions(context.Background(), provider.TransactionRequest{
		ChainID: chain.Ethereum,
		Address: "0xabc",
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}
	if len(page.Transactions) != 1 {
		t.Fatalf("txs = %d", len(page.Transactions))
	}
	if page.NextPageToken == nil || *page.NextPageToken != "cursor1" {
		t.Fatalf("next = %v", page.NextPageToken)
	}
	tx := page.Transactions[0]
	if tx.TxHash != "0xhash" || len(tx.Transfers) != 1 || tx.Transfers[0].Direction != provider.DirectionIn {
		t.Fatalf("tx = %+v", tx)
	}
}
