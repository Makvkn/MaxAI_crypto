// Package coingecko adapts CoinGecko to the market data port. Nothing outside
// this package knows where prices come from (§34, §35).
package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/provider"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/providers/httpx"
)

// Provider implements provider.MarketDataProvider.
type Provider struct {
	client *httpx.Client
}

// New builds the adapter from provider configuration.
func New(cfg config.ProviderConfig) *Provider {
	return &Provider{
		client: httpx.New(provider.CoinGecko, httpx.Config{
			BaseURL:         cfg.CoinGecko.BaseURL,
			APIKey:          cfg.CoinGecko.APIKey,
			Timeout:         cfg.Timeout,
			MaxAttempts:     cfg.MaxAttempts,
			BackoffSchedule: cfg.BackoffSchedule,
		}),
	}
}

// Name implements provider.MarketDataProvider.
func (p *Provider) Name() provider.Name { return provider.CoinGecko }

// GetPrices implements provider.MarketDataProvider.
func (p *Provider) GetPrices(ctx context.Context, req provider.PriceRequest) ([]provider.PriceQuote, error) {
	if len(req.MarketDataIDs) == 0 {
		return nil, nil
	}
	currency := strings.ToLower(string(req.Currency))
	if currency == "" {
		currency = "usd"
	}

	q := url.Values{}
	q.Set("ids", strings.Join(req.MarketDataIDs, ","))
	q.Set("vs_currencies", currency)
	q.Set("include_24hr_change", "true")

	var payload map[string]map[string]json.Number
	if err := p.getJSON(ctx, "/simple/price", q, &payload); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	quotes := make([]provider.PriceQuote, 0, len(payload))
	for id, fields := range payload {
		raw, ok := fields[currency]
		if !ok {
			continue
		}
		price, err := shared.ParseDecimal(raw.String())
		if err != nil {
			continue
		}
		quote := provider.PriceQuote{
			MarketDataID: id,
			Currency:     shared.Currency(strings.ToUpper(currency)),
			Price:        price,
			AsOf:         now,
		}
		if changeRaw, ok := fields[currency+"_24h_change"]; ok {
			if change, err := shared.ParseDecimal(changeRaw.String()); err == nil {
				quote.Change24hPct = shared.Known(change)
			}
		}
		quotes = append(quotes, quote)
	}
	return quotes, nil
}

// GetHistoricalPrices implements provider.MarketDataProvider.
func (p *Provider) GetHistoricalPrices(ctx context.Context, req provider.HistoricalPriceRequest) ([]provider.HistoricalPrice, error) {
	if req.MarketDataID == "" {
		return nil, nil
	}
	currency := strings.ToLower(string(req.Currency))
	if currency == "" {
		currency = "usd"
	}

	q := url.Values{}
	q.Set("vs_currency", currency)
	q.Set("from", fmt.Sprintf("%d", req.From.UTC().Unix()))
	q.Set("to", fmt.Sprintf("%d", req.To.UTC().Unix()))

	var payload struct {
		Prices [][]json.Number `json:"prices"`
	}
	path := "/coins/" + url.PathEscape(req.MarketDataID) + "/market_chart/range"
	if err := p.getJSON(ctx, path, q, &payload); err != nil {
		return nil, err
	}

	points := make([]provider.HistoricalPrice, 0, len(payload.Prices))
	for _, row := range payload.Prices {
		if len(row) < 2 {
			continue
		}
		ms, err := row[0].Int64()
		if err != nil {
			continue
		}
		price, err := shared.ParseDecimal(row[1].String())
		if err != nil {
			continue
		}
		points = append(points, provider.HistoricalPrice{
			MarketDataID: req.MarketDataID,
			Currency:     shared.Currency(strings.ToUpper(currency)),
			Price:        price,
			AsOf:         time.UnixMilli(ms).UTC(),
		})
	}
	return points, nil
}

// ResolveMarketDataID implements provider.MarketDataProvider.
func (p *Provider) ResolveMarketDataID(ctx context.Context, req provider.MappingRequest) (string, bool, error) {
	if req.IsNative || req.ContractAddress == nil || *req.ContractAddress == "" {
		if id, ok := nativeMarketIDs[chain.ID(req.ChainID)]; ok {
			return id, true, nil
		}
		return "", false, nil
	}

	platform, ok := chainPlatforms[chain.ID(req.ChainID)]
	if !ok {
		return "", false, nil
	}

	path := "/coins/" + url.PathEscape(platform) + "/contract/" + url.PathEscape(strings.ToLower(*req.ContractAddress))
	var payload struct {
		ID string `json:"id"`
	}
	if err := p.getJSON(ctx, path, nil, &payload); err != nil {
		if appErr := apperr.From(err); appErr != nil && appErr.Code == apperr.CodeNotFound {
			return "", false, nil
		}
		return "", false, err
	}
	if payload.ID == "" {
		return "", false, nil
	}
	return payload.ID, true, nil
}

var nativeMarketIDs = map[chain.ID]string{
	chain.Ethereum:  "ethereum",
	chain.Bitcoin:   "bitcoin",
	chain.BNBChain:  "binancecoin",
	chain.Solana:    "solana",
	chain.Litecoin:  "litecoin",
	chain.XRPLedger: "ripple",
	chain.Tron:      "tron",
	chain.Dogecoin:  "dogecoin",
}

var chainPlatforms = map[chain.ID]string{
	chain.Ethereum: "ethereum",
	chain.BNBChain: "binance-smart-chain",
	chain.Solana:   "solana",
	chain.Tron:     "tron",
}

func (p *Provider) getJSON(ctx context.Context, path string, query url.Values, dst any) error {
	endpoint := strings.TrimRight(p.client.BaseURL(), "/") + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return apperr.Wrap(apperr.CodeProviderError, err).WithDetail("provider", string(provider.CoinGecko))
	}
	req.Header.Set("Accept", "application/json")
	if key := p.client.APIKey(); key != "" {
		req.Header.Set("x-cg-demo-api-key", key)
		req.Header.Set("x-cg-pro-api-key", key)
	}

	resp, err := p.client.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := httpx.MapStatus(provider.CoinGecko, resp.StatusCode); err != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return err
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return apperr.Wrap(apperr.CodeProviderError, err).WithDetail("provider", string(provider.CoinGecko))
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return apperr.Wrap(apperr.CodeProviderError, err).WithDetail("provider", string(provider.CoinGecko))
	}
	return nil
}

var _ provider.MarketDataProvider = (*Provider)(nil)
