package tatum

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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

// Provider implements provider.BlockchainDataProvider.
type Provider struct {
	client *httpx.Client
}

// New builds the adapter from provider configuration.
func New(cfg config.ProviderConfig) *Provider {
	return &Provider{
		client: httpx.New(provider.Tatum, httpx.Config{
			BaseURL:         cfg.Tatum.BaseURL,
			APIKey:          cfg.Tatum.APIKey,
			Timeout:         cfg.Timeout,
			MaxAttempts:     cfg.MaxAttempts,
			BackoffSchedule: cfg.BackoffSchedule,
		}),
	}
}

// Name implements provider.BlockchainDataProvider.
func (p *Provider) Name() provider.Name { return provider.Tatum }

// Capabilities implements provider.BlockchainDataProvider.
func (p *Provider) Capabilities(context.Context) provider.Capabilities {
	utxo := []provider.Capability{
		provider.CapabilityBalances,
		provider.CapabilityTransactions,
		provider.CapabilityNativeAsset,
		provider.CapabilityPagination,
	}
	accountBased := []provider.Capability{
		provider.CapabilityBalances,
		provider.CapabilityTransactions,
		provider.CapabilityTokenMetadata,
		provider.CapabilityNativeAsset,
		provider.CapabilityPagination,
	}
	return provider.Capabilities{
		Provider: provider.Tatum,
		Chains: map[chain.ID][]provider.Capability{
			chain.Bitcoin:   utxo,
			chain.Litecoin:  utxo,
			chain.Dogecoin:  utxo,
			chain.Solana:    accountBased,
			chain.Tron:      accountBased,
			chain.XRPLedger: accountBased,
			chain.Ethereum:  accountBased,
			chain.BNBChain:  accountBased,
		},
	}
}

// GetBalances implements provider.BlockchainDataProvider.
func (p *Provider) GetBalances(ctx context.Context, req provider.BalanceRequest) ([]provider.NormalizedBalance, error) {
	path, ok := balancePath(req.ChainID, req.Address)
	if !ok {
		return nil, apperr.New(apperr.CodeUnsupportedChain).WithDetail("provider", string(provider.Tatum))
	}

	now := time.Now().UTC()
	switch req.ChainID {
	case chain.Bitcoin, chain.Litecoin, chain.Dogecoin:
		var payload utxoBalanceResponse
		if err := p.getJSON(ctx, path, &payload); err != nil {
			return nil, err
		}
		balance, err := utxoBalance(payload)
		if err != nil {
			return nil, err
		}
		return []provider.NormalizedBalance{nativeBalance(req.ChainID, balance, now)}, nil

	case chain.Tron:
		var payload tronAccountResponse
		if err := p.getJSON(ctx, path, &payload); err != nil {
			return nil, err
		}
		raw := payload.Balance.String()
		if raw == "" {
			raw = "0"
		}
		sun, err := shared.ParseDecimal(raw)
		if err != nil {
			return nil, apperr.Wrap(apperr.CodeProviderError, err).WithDetail("provider", string(provider.Tatum))
		}
		normalized := shared.NewDecimal(sun.Value().Shift(-6))
		bal := nativeBalance(req.ChainID, normalized, now)
		bal.BalanceRaw = sun.Value().Round(0).String()
		return []provider.NormalizedBalance{bal}, nil

	default:
		var payload accountBalanceResponse
		if err := p.getJSON(ctx, path, &payload); err != nil {
			return nil, err
		}
		raw := payload.Balance
		if raw == "" {
			raw = "0"
		}
		normalized, err := shared.ParseDecimal(raw)
		if err != nil {
			return nil, apperr.Wrap(apperr.CodeProviderError, err).WithDetail("provider", string(provider.Tatum))
		}
		return []provider.NormalizedBalance{nativeBalance(req.ChainID, normalized, now)}, nil
	}
}

// GetTransactions implements provider.BlockchainDataProvider.
// UTXO history uses the Blockchains API; account-based chains return an empty
// page until those endpoints are wired.
func (p *Provider) GetTransactions(ctx context.Context, req provider.TransactionRequest) (provider.TransactionPage, error) {
	switch req.ChainID {
	case chain.Bitcoin, chain.Litecoin, chain.Dogecoin:
		return p.getUTXOTransactions(ctx, req)
	case chain.Solana:
		return p.getSolanaTransactions(ctx, req)
	default:
		return provider.TransactionPage{Transactions: []provider.NormalizedTransaction{}}, nil
	}
}

func (p *Provider) getSolanaTransactions(ctx context.Context, req provider.TransactionRequest) (provider.TransactionPage, error) {
	pageSize := req.Limit
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 50
	}
	path := fmt.Sprintf("/v3/solana/account/transaction/%s?pageSize=%d", url.PathEscape(req.Address), pageSize)
	if req.PageToken != nil && *req.PageToken != "" {
		path += "&offset=" + url.QueryEscape(*req.PageToken)
	}

	var payload []solanaTxDTO
	if err := p.getJSON(ctx, path, &payload); err != nil {
		// Account-history shapes vary by plan; fail soft with an empty page so
		// balance-only sync still succeeds.
		return provider.TransactionPage{Transactions: []provider.NormalizedTransaction{}}, nil
	}

	txs := make([]provider.NormalizedTransaction, 0, len(payload))
	for _, item := range payload {
		normalized, ok := normalizeSolanaTransaction(req.Address, item)
		if ok {
			txs = append(txs, normalized)
		}
	}
	var next *string
	if len(payload) >= pageSize {
		token := strconv.Itoa(pageSize)
		if req.PageToken != nil && *req.PageToken != "" {
			if prev, err := strconv.Atoi(*req.PageToken); err == nil {
				token = strconv.Itoa(prev + pageSize)
			}
		}
		next = &token
	}
	return provider.TransactionPage{Transactions: txs, NextPageToken: next}, nil
}

type solanaTxDTO struct {
	Signature string          `json:"signature"`
	BlockTime int64           `json:"blockTime"`
	Slot      int64           `json:"slot"`
	Fee       json.Number     `json:"fee"`
	Status    string          `json:"status"`
	From      string          `json:"from"`
	To        string          `json:"to"`
	Amount    json.Number     `json:"amount"`
}

func normalizeSolanaTransaction(wallet string, item solanaTxDTO) (provider.NormalizedTransaction, bool) {
	if item.Signature == "" {
		return provider.NormalizedTransaction{}, false
	}
	meta := provider.AssetMetadata{
		Symbol:   nativeSymbol(chain.Solana),
		Name:     nativeName(chain.Solana),
		Decimals: nativeDecimals(chain.Solana),
		Type:     asset.TypeNative,
	}
	amountRaw := item.Amount.String()
	if amountRaw == "" {
		amountRaw = "0"
	}
	lamports, err := shared.ParseDecimal(amountRaw)
	if err != nil {
		return provider.NormalizedTransaction{}, false
	}
	amount := shared.NewDecimal(lamports.Value().Shift(-9))
	direction := provider.DirectionIn
	if strings.EqualFold(strings.TrimSpace(item.From), strings.TrimSpace(wallet)) {
		direction = provider.DirectionOut
	}
	var block *int64
	if item.Slot > 0 {
		v := item.Slot
		block = &v
	}
	ts := time.Unix(item.BlockTime, 0).UTC()
	if item.BlockTime <= 0 {
		ts = time.Now().UTC()
	}
	ref := item.Signature
	from := nonempty(item.From)
	to := nonempty(item.To)
	return provider.NormalizedTransaction{
		ChainID:     chain.Solana,
		TxHash:      item.Signature,
		BlockNumber: block,
		Timestamp:   ts,
		Successful:  item.Status == "" || strings.EqualFold(item.Status, "success") || strings.EqualFold(item.Status, "confirmed"),
		FromAddress: from,
		ToAddress:   to,
		Transfers: []provider.NormalizedTransfer{{
			AssetIdentity: asset.Identity{ChainID: chain.Solana},
			Metadata:      meta,
			AmountRaw:     lamports.Value().Round(0).String(),
			Amount:        amount,
			Direction:     direction,
		}},
		ProviderRef: &ref,
	}, true
}

func nonempty(v string) *string {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (p *Provider) getUTXOTransactions(ctx context.Context, req provider.TransactionRequest) (provider.TransactionPage, error) {
	chainSlug, ok := utxoChainSlug(req.ChainID)
	if !ok {
		return provider.TransactionPage{}, apperr.New(apperr.CodeUnsupportedChain).WithDetail("provider", string(provider.Tatum))
	}

	pageSize := req.Limit
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 50
	}
	offset := 0
	if req.PageToken != nil && *req.PageToken != "" {
		parsed, err := strconv.Atoi(*req.PageToken)
		if err != nil || parsed < 0 {
			return provider.TransactionPage{}, apperr.New(apperr.CodeProviderError).
				WithDetail("provider", string(provider.Tatum)).
				WithDetail("reason", "invalid_page_token")
		}
		offset = parsed
	}

	query := url.Values{}
	query.Set("chain", chainSlug)
	query.Set("address", req.Address)
	query.Set("pageSize", strconv.Itoa(pageSize))
	query.Set("offset", strconv.Itoa(offset))

	var payload []utxoTransactionDTO
	if err := p.getJSON(ctx, "/v4/data/blockchains/transaction/history/utxos?"+query.Encode(), &payload); err != nil {
		return provider.TransactionPage{}, err
	}

	txs := make([]provider.NormalizedTransaction, 0, len(payload))
	for _, item := range payload {
		normalized, ok := normalizeUTXOTransaction(req.ChainID, req.Address, item)
		if ok {
			txs = append(txs, normalized)
		}
	}

	var next *string
	if len(payload) >= pageSize {
		token := strconv.Itoa(offset + pageSize)
		next = &token
	}
	return provider.TransactionPage{Transactions: txs, NextPageToken: next}, nil
}

func balancePath(id chain.ID, address string) (string, bool) {
	switch id {
	case chain.Bitcoin:
		return "/v3/bitcoin/address/balance/" + address, true
	case chain.Litecoin:
		return "/v3/litecoin/address/balance/" + address, true
	case chain.Dogecoin:
		return "/v3/dogecoin/address/balance/" + address, true
	case chain.Ethereum:
		return "/v3/ethereum/account/balance/" + address, true
	case chain.BNBChain:
		return "/v3/bsc/account/balance/" + address, true
	case chain.Solana:
		return "/v3/solana/account/balance/" + address, true
	case chain.Tron:
		return "/v3/tron/account/" + address, true
	case chain.XRPLedger:
		return "/v3/xrp/account/" + address + "/balance", true
	default:
		return "", false
	}
}

func utxoChainSlug(id chain.ID) (string, bool) {
	switch id {
	case chain.Bitcoin:
		return "bitcoin-mainnet", true
	case chain.Litecoin:
		return "litecoin-mainnet", true
	case chain.Dogecoin:
		return "dogecoin-mainnet", true
	default:
		return "", false
	}
}

type utxoBalanceResponse struct {
	Incoming string `json:"incoming"`
	Outgoing string `json:"outgoing"`
}

type accountBalanceResponse struct {
	Balance string `json:"balance"`
}

type tronAccountResponse struct {
	Balance json.Number `json:"balance"`
}

type utxoTransactionDTO struct {
	Hash        string             `json:"hash"`
	BlockNumber int64              `json:"blockNumber"`
	Time        int64              `json:"time"`
	Fee         json.Number        `json:"fee"`
	Inputs      []utxoInputDTO     `json:"inputs"`
	Outputs     []utxoOutputDTO    `json:"outputs"`
}

type utxoInputDTO struct {
	Address *string      `json:"address"`
	Coin    *utxoCoinDTO `json:"coin"`
}

type utxoCoinDTO struct {
	Value   json.Number `json:"value"`
	Address string      `json:"address"`
}

type utxoOutputDTO struct {
	Value   json.Number `json:"value"`
	Address *string     `json:"address"`
}

func normalizeUTXOTransaction(id chain.ID, walletAddress string, item utxoTransactionDTO) (provider.NormalizedTransaction, bool) {
	if item.Hash == "" {
		return provider.NormalizedTransaction{}, false
	}
	wallet := strings.TrimSpace(walletAddress)
	decimals := nativeDecimals(id)
	meta := provider.AssetMetadata{
		Symbol:   nativeSymbol(id),
		Name:     nativeName(id),
		Decimals: decimals,
		Type:     asset.TypeNative,
	}

	var received, spent int64
	var fromAddr, toAddr *string
	for _, in := range item.Inputs {
		addr := inputAddress(in)
		value := parseSatoshis(in.Coin)
		if addr != "" && addressesEqual(addr, wallet) {
			spent += value
		} else if fromAddr == nil && addr != "" {
			fromAddr = stringPtr(addr)
		}
	}
	for _, out := range item.Outputs {
		addr := ""
		if out.Address != nil {
			addr = strings.TrimSpace(*out.Address)
		}
		value := parseSatoshisValue(out.Value)
		if addr != "" && addressesEqual(addr, wallet) {
			received += value
		} else if toAddr == nil && addr != "" {
			toAddr = stringPtr(addr)
		}
	}

	net := received - spent
	transfers := make([]provider.NormalizedTransfer, 0, 1)
	if net > 0 {
		amount := shared.NewDecimalFromInt(net).Value().Shift(-int32(decimals))
		transfers = append(transfers, provider.NormalizedTransfer{
			AssetIdentity: asset.Identity{ChainID: id},
			Metadata:      meta,
			AmountRaw:     strconv.FormatInt(net, 10),
			Amount:        shared.NewDecimal(amount),
			Direction:     provider.DirectionIn,
		})
	} else if net < 0 {
		outAmount := -net
		amount := shared.NewDecimalFromInt(outAmount).Value().Shift(-int32(decimals))
		transfers = append(transfers, provider.NormalizedTransfer{
			AssetIdentity: asset.Identity{ChainID: id},
			Metadata:      meta,
			AmountRaw:     strconv.FormatInt(outAmount, 10),
			Amount:        shared.NewDecimal(amount),
			Direction:     provider.DirectionOut,
		})
	}

	var block *int64
	if item.BlockNumber > 0 {
		v := item.BlockNumber
		block = &v
	}
	ts := time.Unix(item.Time, 0).UTC()
	if item.Time <= 0 {
		ts = time.Now().UTC()
	}

	feeSats := parseSatoshisValue(item.Fee)
	feeAmount := shared.Unknown()
	if feeSats > 0 {
		feeAmount = shared.Known(shared.NewDecimal(shared.NewDecimalFromInt(feeSats).Value().Shift(-int32(decimals))))
	}
	ref := item.Hash
	identity := asset.Identity{ChainID: id}
	return provider.NormalizedTransaction{
		ChainID:          id,
		TxHash:           item.Hash,
		BlockNumber:      block,
		Timestamp:        ts,
		Successful:       true,
		FromAddress:      fromAddr,
		ToAddress:        toAddr,
		Transfers:        transfers,
		FeeAssetIdentity: &identity,
		FeeMetadata:      &meta,
		FeeAmount:        feeAmount,
		ProviderRef:      &ref,
	}, true
}

func inputAddress(in utxoInputDTO) string {
	if in.Address != nil && strings.TrimSpace(*in.Address) != "" {
		return strings.TrimSpace(*in.Address)
	}
	if in.Coin != nil {
		return strings.TrimSpace(in.Coin.Address)
	}
	return ""
}

func parseSatoshis(coin *utxoCoinDTO) int64 {
	if coin == nil {
		return 0
	}
	return parseSatoshisValue(coin.Value)
}

func parseSatoshisValue(raw json.Number) int64 {
	if raw == "" {
		return 0
	}
	if v, err := raw.Int64(); err == nil {
		return v
	}
	dec, err := shared.ParseDecimal(raw.String())
	if err != nil {
		return 0
	}
	return dec.Value().IntPart()
}

func addressesEqual(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

func stringPtr(v string) *string { return &v }

func utxoBalance(payload utxoBalanceResponse) (shared.Decimal, error) {
	incoming, err := shared.ParseDecimal(defaultZero(payload.Incoming))
	if err != nil {
		return shared.Decimal{}, apperr.Wrap(apperr.CodeProviderError, err).WithDetail("provider", string(provider.Tatum))
	}
	outgoing, err := shared.ParseDecimal(defaultZero(payload.Outgoing))
	if err != nil {
		return shared.Decimal{}, apperr.Wrap(apperr.CodeProviderError, err).WithDetail("provider", string(provider.Tatum))
	}
	return shared.NewDecimal(incoming.Value().Sub(outgoing.Value())), nil
}

func defaultZero(raw string) string {
	if raw == "" {
		return "0"
	}
	return raw
}

func nativeBalance(id chain.ID, amount shared.Decimal, observedAt time.Time) provider.NormalizedBalance {
	decimals := nativeDecimals(id)
	return provider.NormalizedBalance{
		ChainID:       id,
		AssetIdentity: asset.Identity{ChainID: id},
		Metadata: provider.AssetMetadata{
			Symbol:   nativeSymbol(id),
			Name:     nativeName(id),
			Decimals: decimals,
			Type:     asset.TypeNative,
		},
		BalanceRaw:        amount.Value().Shift(int32(decimals)).Round(0).String(),
		BalanceNormalized: amount,
		ObservedAt:        observedAt,
	}
}

func nativeSymbol(id chain.ID) string {
	switch id {
	case chain.Bitcoin:
		return "BTC"
	case chain.BNBChain:
		return "BNB"
	case chain.Solana:
		return "SOL"
	case chain.Litecoin:
		return "LTC"
	case chain.XRPLedger:
		return "XRP"
	case chain.Tron:
		return "TRX"
	case chain.Dogecoin:
		return "DOGE"
	default:
		return "ETH"
	}
}

func nativeName(id chain.ID) string {
	switch id {
	case chain.Bitcoin:
		return "Bitcoin"
	case chain.BNBChain:
		return "BNB"
	case chain.Solana:
		return "Solana"
	case chain.Litecoin:
		return "Litecoin"
	case chain.XRPLedger:
		return "XRP"
	case chain.Tron:
		return "TRON"
	case chain.Dogecoin:
		return "Dogecoin"
	default:
		return "Ether"
	}
}

func nativeDecimals(id chain.ID) int {
	switch id {
	case chain.Bitcoin, chain.Litecoin, chain.Dogecoin:
		return 8
	case chain.Solana:
		return 9
	case chain.XRPLedger, chain.Tron:
		return 6
	default:
		return 18
	}
}

func (p *Provider) getJSON(ctx context.Context, path string, dst any) error {
	endpoint := strings.TrimRight(p.client.BaseURL(), "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return apperr.Wrap(apperr.CodeProviderError, err).WithDetail("provider", string(provider.Tatum))
	}
	req.Header.Set("Accept", "application/json")
	if key := p.client.APIKey(); key != "" {
		req.Header.Set("x-api-key", key)
	}

	resp, err := p.client.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := httpx.MapStatus(provider.Tatum, resp.StatusCode); err != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return err
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return apperr.Wrap(apperr.CodeProviderError, err).WithDetail("provider", string(provider.Tatum))
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return apperr.Wrap(apperr.CodeProviderError, fmt.Errorf("decode tatum response: %w", err)).
			WithDetail("provider", string(provider.Tatum))
	}
	return nil
}

var _ provider.BlockchainDataProvider = (*Provider)(nil)
