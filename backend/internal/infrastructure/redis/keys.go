package redis

import (
	"fmt"
	"time"
)

// Key namespaces from §64. Building keys here rather than inline keeps the
// keyspace reviewable and prevents accidental collisions between subsystems.
const (
	nsPrice     = "price"
	nsPortfolio = "portfolio"
	nsAIUsage   = "ai:usage"
	nsSyncLock  = "sync:wallet"
	nsRateLimit = "ratelimit"
)

// PriceKey is the cache key for a single asset price.
func PriceKey(assetID string) string {
	return fmt.Sprintf("%s:%s", nsPrice, assetID)
}

// PortfolioKey is the cache key for a wallet's computed portfolio.
func PortfolioKey(walletID string) string {
	return fmt.Sprintf("%s:%s", nsPortfolio, walletID)
}

// AIUsageKey is the daily AI usage counter for a user. The date component uses
// UTC day boundaries (§86).
func AIUsageKey(userID string, day time.Time) string {
	return fmt.Sprintf("%s:%s:%s", nsAIUsage, userID, day.UTC().Format(time.DateOnly))
}

// WalletSyncLockKey is the mutual-exclusion key preventing two concurrent full
// synchronizations of the same wallet (§61).
func WalletSyncLockKey(walletID string) string {
	return fmt.Sprintf("%s:%s", nsSyncLock, walletID)
}

// RateLimitKey scopes a rate-limit counter to a limiter layer and subject,
// keeping the IP, anonymous and authenticated layers independent (§153).
func RateLimitKey(scope, subject string, window time.Time) string {
	return fmt.Sprintf("%s:%s:%s:%d", nsRateLimit, scope, subject, window.UTC().Unix())
}

// NextUTCMidnight returns the TTL that expires the AI usage counter at the next
// UTC day boundary (§65).
func NextUTCMidnight(now time.Time) time.Duration {
	utc := now.UTC()
	midnight := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)
	return midnight.Sub(utc)
}
