package pricing

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/price"
	"github.com/maxaicrypto/backend/internal/domain/provider"
	"github.com/maxaicrypto/backend/internal/domain/shared"
)

// devQuotes are deterministic USD prices used when market providers are not
// configured. They exist only to complete the first vertical slice locally.
var devQuotes = map[string]string{
	"ethereum":    "2500",
	"bitcoin":     "60000",
	"binancecoin": "400",
	"solana":      "150",
	"litecoin":    "80",
	"ripple":      "0.6",
	"tron":        "0.12",
	"dogecoin":    "0.15",
}

// App implements Service.
type App struct {
	assets asset.Repository
	prices price.Repository
	market provider.Resolver
	useDev bool
}

// NewApp wires the pricing service.
func NewApp(assets asset.Repository, prices price.Repository, market provider.Resolver, useDevQuotes bool) *App {
	return &App{assets: assets, prices: prices, market: market, useDev: useDevQuotes}
}

// GetCurrent implements Service.
func (a *App) GetCurrent(ctx context.Context, assetIDs []uuid.UUID) (map[uuid.UUID]price.Price, error) {
	return a.prices.GetLatest(ctx, assetIDs)
}

// GetAt implements Service.
func (a *App) GetAt(ctx context.Context, assetID uuid.UUID, at time.Time) (price.Price, bool, error) {
	return a.prices.GetClosest(ctx, assetID, at)
}

// Refresh implements Service.
func (a *App) Refresh(ctx context.Context, assetIDs []uuid.UUID) (int, error) {
	if len(assetIDs) == 0 {
		return 0, nil
	}
	assets, err := a.assets.GetManyByID(ctx, assetIDs)
	if err != nil {
		return 0, err
	}

	now := time.Now().UTC()
	if a.useDev {
		return a.refreshDev(ctx, assets, now)
	}
	return a.refreshMarket(ctx, assets, now)
}

// ResolveMapping implements Service.
func (a *App) ResolveMapping(ctx context.Context, ast asset.Asset) (bool, error) {
	if ast.HasMarketData() {
		return true, nil
	}

	if a.useDev {
		id, ok := nativeDevID(ast)
		if !ok {
			return false, nil
		}
		providerName := asset.MarketDataCoinGecko
		if err := a.assets.SetMarketDataMapping(ctx, ast.ID, &providerName, &id); err != nil {
			return false, err
		}
		return true, nil
	}

	market, err := a.market.ResolveMarket(ctx)
	if err != nil {
		return false, err
	}
	marketID, ok, err := market.ResolveMarketDataID(ctx, provider.MappingRequest{
		ChainID:         string(ast.ChainID),
		ContractAddress: ast.ContractAddress,
		Symbol:          ast.Symbol,
		IsNative:        ast.Type == asset.TypeNative || ast.ContractAddress == nil,
	})
	if err != nil || !ok {
		return false, err
	}
	providerName := asset.MarketDataCoinGecko
	if err := a.assets.SetMarketDataMapping(ctx, ast.ID, &providerName, &marketID); err != nil {
		return false, err
	}
	return true, nil
}

func (a *App) refreshDev(ctx context.Context, assets map[uuid.UUID]asset.Asset, now time.Time) (int, error) {
	refreshed := 0
	for id, ast := range assets {
		if ast.MarketDataID == nil {
			continue
		}
		raw, ok := devQuotes[*ast.MarketDataID]
		if !ok {
			continue
		}
		value, err := shared.ParseDecimal(raw)
		if err != nil {
			return refreshed, apperr.Wrap(apperr.CodeInternal, err)
		}
		if err := a.prices.UpsertLatest(ctx, price.Price{
			AssetID:  id,
			Currency: shared.CurrencyUSD,
			ValueUSD: shared.Known(value),
			Status:   price.StatusAvailable,
			Source:   price.SourceCoinGecko,
			AsOf:     now,
		}); err != nil {
			return refreshed, err
		}
		refreshed++
	}
	return refreshed, nil
}

func (a *App) refreshMarket(ctx context.Context, assets map[uuid.UUID]asset.Asset, now time.Time) (int, error) {
	market, err := a.market.ResolveMarket(ctx)
	if err != nil {
		return 0, err
	}

	ids := make([]string, 0, len(assets))
	byMarketID := make(map[string][]uuid.UUID, len(assets))
	for id, ast := range assets {
		if ast.MarketDataID == nil || *ast.MarketDataID == "" {
			continue
		}
		marketID := *ast.MarketDataID
		if _, seen := byMarketID[marketID]; !seen {
			ids = append(ids, marketID)
		}
		byMarketID[marketID] = append(byMarketID[marketID], id)
	}
	if len(ids) == 0 {
		return 0, nil
	}

	quotes, err := market.GetPrices(ctx, provider.PriceRequest{
		MarketDataIDs: ids,
		Currency:      shared.CurrencyUSD,
	})
	if err != nil {
		return 0, err
	}

	refreshed := 0
	for _, quote := range quotes {
		assetIDs := byMarketID[quote.MarketDataID]
		for _, assetID := range assetIDs {
			p := price.Price{
				AssetID:  assetID,
				Currency: shared.CurrencyUSD,
				ValueUSD: shared.Known(quote.Price),
				Status:   price.StatusAvailable,
				Source:   price.SourceCoinGecko,
				AsOf:     quote.AsOf,
			}
			if p.AsOf.IsZero() {
				p.AsOf = now
			}
			if quote.Change24hPct.Valid {
				p.Change24h = quote.Change24hPct
			}
			if err := a.prices.UpsertLatest(ctx, p); err != nil {
				return refreshed, err
			}
			refreshed++
		}
	}
	return refreshed, nil
}

func nativeDevID(ast asset.Asset) (string, bool) {
	if ast.Type != asset.TypeNative && ast.ContractAddress != nil {
		return "", false
	}
	switch string(ast.ChainID) {
	case "ethereum":
		return "ethereum", true
	case "bitcoin":
		return "bitcoin", true
	case "bnb":
		return "binancecoin", true
	case "solana":
		return "solana", true
	case "litecoin":
		return "litecoin", true
	case "xrpl":
		return "ripple", true
	case "tron":
		return "tron", true
	case "dogecoin":
		return "dogecoin", true
	default:
		return "", false
	}
}

var _ Service = (*App)(nil)
