// Package zerion adapts the Zerion API to the blockchain data port. Zerion DTOs
// exist only inside this package; callers receive normalized domain types
// (§22, §29).
package zerion

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/provider"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/providers/httpx"
)

// Provider implements provider.BlockchainDataProvider for EVM chains.
type Provider struct {
	client *httpx.Client
}

// New builds the adapter from provider configuration.
func New(cfg config.ProviderConfig) *Provider {
	return &Provider{
		client: httpx.New(provider.Zerion, httpx.Config{
			BaseURL:         cfg.Zerion.BaseURL,
			APIKey:          cfg.Zerion.APIKey,
			Timeout:         cfg.Timeout,
			MaxAttempts:     cfg.MaxAttempts,
			BackoffSchedule: cfg.BackoffSchedule,
		}),
	}
}

// Name implements provider.BlockchainDataProvider.
func (p *Provider) Name() provider.Name { return provider.Zerion }

// Capabilities implements provider.BlockchainDataProvider. Zerion serves the
// EVM chains of the MVP set (§20, §23).
func (p *Provider) Capabilities(context.Context) provider.Capabilities {
	evm := []provider.Capability{
		provider.CapabilityBalances,
		provider.CapabilityTransactions,
		provider.CapabilityTokenMetadata,
		provider.CapabilityNativeAsset,
		provider.CapabilityPagination,
	}
	return provider.Capabilities{
		Provider: provider.Zerion,
		Chains: map[chain.ID][]provider.Capability{
			chain.Ethereum: evm,
			chain.BNBChain: evm,
		},
	}
}

// GetBalances implements provider.BlockchainDataProvider.
func (p *Provider) GetBalances(ctx context.Context, req provider.BalanceRequest) ([]provider.NormalizedBalance, error) {
	zerionChain, ok := toZerionChain(req.ChainID)
	if !ok {
		return nil, apperr.New(apperr.CodeUnsupportedChain).WithDetail("provider", string(provider.Zerion))
	}

	q := url.Values{}
	q.Set("filter[chain_ids]", zerionChain)
	q.Set("filter[positions]", "only_simple")
	q.Set("filter[trash]", "only_non_trash")
	q.Set("currency", "usd")

	path := "/v1/wallets/" + url.PathEscape(req.Address) + "/positions/"
	var payload positionsResponse
	if err := p.getJSON(ctx, path, q, &payload); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	balances := make([]provider.NormalizedBalance, 0, len(payload.Data))
	for _, item := range payload.Data {
		balance, ok := normalizePosition(req.ChainID, zerionChain, item, now)
		if !ok {
			continue
		}
		balances = append(balances, balance)
	}
	return balances, nil
}

// GetTransactions implements provider.BlockchainDataProvider.
func (p *Provider) GetTransactions(ctx context.Context, req provider.TransactionRequest) (provider.TransactionPage, error) {
	zerionChain, ok := toZerionChain(req.ChainID)
	if !ok {
		return provider.TransactionPage{}, apperr.New(apperr.CodeUnsupportedChain).WithDetail("provider", string(provider.Zerion))
	}

	q := url.Values{}
	q.Set("filter[chain_ids]", zerionChain)
	q.Set("currency", "usd")
	limit := req.Limit
	if limit < 1 {
		limit = 50
	}
	q.Set("page[size]", fmt.Sprintf("%d", limit))
	if req.PageToken != nil && *req.PageToken != "" {
		q.Set("page[after]", *req.PageToken)
	}
	if req.Since != nil {
		q.Set("filter[min_mined_at]", fmt.Sprintf("%d", req.Since.UTC().UnixMilli()))
	}

	path := "/v1/wallets/" + url.PathEscape(req.Address) + "/transactions/"
	var payload transactionsResponse
	if err := p.getJSON(ctx, path, q, &payload); err != nil {
		return provider.TransactionPage{}, err
	}

	txs := make([]provider.NormalizedTransaction, 0, len(payload.Data))
	for _, item := range payload.Data {
		tx, ok := normalizeTransaction(req.ChainID, zerionChain, item)
		if !ok {
			continue
		}
		txs = append(txs, tx)
	}

	page := provider.TransactionPage{Transactions: txs}
	if payload.Links.Next != "" {
		if after := pageAfter(payload.Links.Next); after != "" {
			page.NextPageToken = &after
		}
	}
	return page, nil
}

func toZerionChain(id chain.ID) (string, bool) {
	switch id {
	case chain.Ethereum:
		return "ethereum", true
	case chain.BNBChain:
		return "binance-smart-chain", true
	default:
		return "", false
	}
}

func normalizePosition(domainChain chain.ID, zerionChain string, item positionResource, now time.Time) (provider.NormalizedBalance, bool) {
	attrs := item.Attributes
	if attrs.Quantity.Int == "" {
		return provider.NormalizedBalance{}, false
	}
	impl := matchingImplementation(attrs.FungibleInfo.Implementations, zerionChain)
	decimals := attrs.Quantity.Decimals
	if impl != nil && impl.Decimals > 0 {
		decimals = impl.Decimals
	}

	normalized, err := shared.ParseDecimal(attrs.Quantity.Numeric)
	if err != nil {
		return provider.NormalizedBalance{}, false
	}

	var contract *string
	assetType := asset.TypeNative
	if impl != nil && impl.Address != nil && *impl.Address != "" {
		addr := strings.ToLower(*impl.Address)
		contract = &addr
		assetType = asset.TypeToken
	}

	symbol := attrs.FungibleInfo.Symbol
	if symbol == "" {
		symbol = "UNKNOWN"
	}
	name := attrs.FungibleInfo.Name
	if name == "" {
		name = symbol
	}

	var iconURL *string
	if attrs.FungibleInfo.Icon != nil && attrs.FungibleInfo.Icon.URL != "" {
		u := attrs.FungibleInfo.Icon.URL
		iconURL = &u
	}

	observedAt := now
	if attrs.UpdatedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, attrs.UpdatedAt); err == nil {
			observedAt = parsed.UTC()
		}
	}

	ref := item.ID
	return provider.NormalizedBalance{
		ChainID: domainChain,
		AssetIdentity: asset.Identity{
			ChainID:         domainChain,
			ContractAddress: contract,
		},
		Metadata: provider.AssetMetadata{
			Symbol:   symbol,
			Name:     name,
			Decimals: decimals,
			Type:     assetType,
			IconURL:  iconURL,
		},
		BalanceRaw:        attrs.Quantity.Int,
		BalanceNormalized: normalized,
		ProviderRef:       &ref,
		ObservedAt:        observedAt,
	}, true
}

func normalizeTransaction(domainChain chain.ID, zerionChain string, item transactionResource) (provider.NormalizedTransaction, bool) {
	attrs := item.Attributes
	if attrs.Hash == "" {
		return provider.NormalizedTransaction{}, false
	}
	minedAt, err := time.Parse(time.RFC3339, attrs.MinedAt)
	if err != nil {
		return provider.NormalizedTransaction{}, false
	}

	var block *int64
	if attrs.MinedAtBlock > 0 {
		v := int64(attrs.MinedAtBlock)
		block = &v
	}

	transfers := make([]provider.NormalizedTransfer, 0, len(attrs.Transfers))
	for _, transfer := range attrs.Transfers {
		normalized, ok := normalizeTransfer(domainChain, zerionChain, transfer)
		if ok {
			transfers = append(transfers, normalized)
		}
	}

	ref := item.ID
	return provider.NormalizedTransaction{
		ChainID:      domainChain,
		TxHash:       attrs.Hash,
		BlockNumber:  block,
		Timestamp:    minedAt.UTC(),
		Successful:   attrs.Status == "confirmed",
		FromAddress:  nonemptyPtr(attrs.SentFrom),
		ToAddress:    nonemptyPtr(attrs.SentTo),
		Transfers:    transfers,
		Protocol:     zerionProtocol(attrs),
		ProviderRef:  &ref,
	}, true
}

func zerionProtocol(attrs transactionAttributes) *string {
	if attrs.ApplicationMetadata != nil && strings.TrimSpace(attrs.ApplicationMetadata.Name) != "" {
		name := strings.TrimSpace(attrs.ApplicationMetadata.Name)
		if op := strings.TrimSpace(attrs.OperationType); op != "" {
			combined := name + "-" + op
			return &combined
		}
		return &name
	}
	if op := strings.TrimSpace(attrs.OperationType); op != "" {
		return &op
	}
	return nil
}

func normalizeTransfer(domainChain chain.ID, zerionChain string, transfer transferDTO) (provider.NormalizedTransfer, bool) {
	if transfer.Quantity.Int == "" {
		return provider.NormalizedTransfer{}, false
	}
	amount, err := shared.ParseDecimal(transfer.Quantity.Numeric)
	if err != nil {
		return provider.NormalizedTransfer{}, false
	}
	impl := matchingImplementation(transfer.FungibleInfo.Implementations, zerionChain)
	decimals := transfer.Quantity.Decimals
	var contract *string
	assetType := asset.TypeNative
	if impl != nil {
		if impl.Decimals > 0 {
			decimals = impl.Decimals
		}
		if impl.Address != nil && *impl.Address != "" {
			addr := strings.ToLower(*impl.Address)
			contract = &addr
			assetType = asset.TypeToken
		}
	}
	direction := provider.DirectionIn
	if transfer.Direction == "out" {
		direction = provider.DirectionOut
	}
	symbol := transfer.FungibleInfo.Symbol
	if symbol == "" {
		symbol = "UNKNOWN"
	}
	name := transfer.FungibleInfo.Name
	if name == "" {
		name = symbol
	}
	return provider.NormalizedTransfer{
		AssetIdentity: asset.Identity{ChainID: domainChain, ContractAddress: contract},
		Metadata: provider.AssetMetadata{
			Symbol:   symbol,
			Name:     name,
			Decimals: decimals,
			Type:     assetType,
		},
		AmountRaw: transfer.Quantity.Int,
		Amount:    amount,
		Direction: direction,
	}, true
}

func matchingImplementation(implementations []implementationDTO, zerionChain string) *implementationDTO {
	for i := range implementations {
		if implementations[i].ChainID == zerionChain {
			return &implementations[i]
		}
	}
	if len(implementations) == 1 {
		return &implementations[0]
	}
	return nil
}

func nonemptyPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func pageAfter(next string) string {
	u, err := url.Parse(next)
	if err != nil {
		return ""
	}
	return u.Query().Get("page[after]")
}

func (p *Provider) getJSON(ctx context.Context, path string, query url.Values, dst any) error {
	endpoint := strings.TrimRight(p.client.BaseURL(), "/") + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return apperr.Wrap(apperr.CodeProviderError, err).WithDetail("provider", string(provider.Zerion))
	}
	req.Header.Set("Accept", "application/json")
	if key := p.client.APIKey(); key != "" {
		token := base64.StdEncoding.EncodeToString([]byte(key + ":"))
		req.Header.Set("Authorization", "Basic "+token)
	}

	resp, err := p.client.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := httpx.MapStatus(provider.Zerion, resp.StatusCode); err != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return err
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return apperr.Wrap(apperr.CodeProviderError, err).WithDetail("provider", string(provider.Zerion))
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return apperr.Wrap(apperr.CodeProviderError, err).WithDetail("provider", string(provider.Zerion))
	}
	return nil
}

var _ provider.BlockchainDataProvider = (*Provider)(nil)
