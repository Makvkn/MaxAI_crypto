package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/maxaicrypto/backend/internal/domain/usage"
)

var reserveScript = redis.NewScript(`
local current = tonumber(redis.call('GET', KEYS[1]) or '0')
local limit = tonumber(ARGV[1])
if current >= limit then
  return -1
end
local next = redis.call('INCR', KEYS[1])
if next == 1 then
  redis.call('EXPIRE', KEYS[1], ARGV[2])
end
if next > limit then
  redis.call('DECR', KEYS[1])
  return -1
end
return next
`)

// UsageCounter implements usage.Counter backed by Redis.
type UsageCounter struct {
	client *Client
}

// NewUsageCounter builds the Redis AI usage counter.
func NewUsageCounter(client *Client) *UsageCounter {
	return &UsageCounter{client: client}
}

// Reserve implements usage.Counter.
func (c *UsageCounter) Reserve(ctx context.Context, userID uuid.UUID, day time.Time, limit int) (usage.Reservation, bool, error) {
	key := AIUsageKey(userID.String(), day)
	ttl := int64(NextUTCMidnight(time.Now().UTC()).Seconds())
	if ttl < 1 {
		ttl = 1
	}

	reservation := usage.Reservation{
		ID:         uuid.New(),
		UserID:     userID,
		ReservedAt: time.Now().UTC(),
	}

	result, err := reserveScript.Run(ctx, c.client, []string{key}, limit, ttl).Int64()
	if err != nil {
		return usage.Reservation{}, false, fmt.Errorf("reserve ai usage: %w", err)
	}
	if result < 0 {
		return usage.Reservation{}, false, nil
	}
	return reservation, true, nil
}

// Commit implements usage.Counter.
func (c *UsageCounter) Commit(ctx context.Context, reservation usage.Reservation) error {
	_ = ctx
	_ = reservation
	return nil
}

// Release implements usage.Counter.
func (c *UsageCounter) Release(ctx context.Context, reservation usage.Reservation) error {
	key := AIUsageKey(reservation.UserID.String(), utcDay(reservation.ReservedAt))
	current, err := c.client.Decr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("release ai usage: %w", err)
	}
	if current < 0 {
		_ = c.client.Set(ctx, key, 0, NextUTCMidnight(time.Now().UTC())).Err()
	}
	return nil
}

// Used implements usage.Counter.
func (c *UsageCounter) Used(ctx context.Context, userID uuid.UUID, day time.Time) (int, error) {
	key := AIUsageKey(userID.String(), day)
	value, err := c.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read ai usage: %w", err)
	}
	if value < 0 {
		return 0, nil
	}
	return value, nil
}

func utcDay(day time.Time) time.Time {
	utc := day.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

var _ usage.Counter = (*UsageCounter)(nil)
