package redis

import (
	"context"
	"time"
)

// RateLimiter enforces the distributed rate-limit layers from §153.
type RateLimiter struct {
	client *Client
}

// NewRateLimiter builds a Redis-backed limiter.
func NewRateLimiter(client *Client) *RateLimiter {
	return &RateLimiter{client: client}
}

// Allow increments the counter for scope/subject in the current UTC minute and
// returns false when the limit would be exceeded.
func (l *RateLimiter) Allow(ctx context.Context, scope, subject string, limitPerMinute int) (bool, error) {
	if limitPerMinute <= 0 {
		return true, nil
	}
	window := time.Now().UTC().Truncate(time.Minute)
	key := RateLimitKey(scope, subject, window)
	n, err := l.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if n == 1 {
		_ = l.client.Expire(ctx, key, 2*time.Minute).Err()
	}
	return n <= int64(limitPerMinute), nil
}
