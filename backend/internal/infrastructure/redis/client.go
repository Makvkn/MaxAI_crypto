// Package redis provides the shared Redis client used for caching, rate
// limiting, distributed locks and job coordination. Redis is infrastructure and
// is never the source of truth (§64).
package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/maxaicrypto/backend/internal/app/config"
)

// Client wraps the Redis client with project configuration and health checks.
type Client struct {
	*goredis.Client
}

// NewClient dials Redis and verifies connectivity before returning.
func NewClient(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	opts, err := goredis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse REDIS_URL: %w", err)
	}
	opts.DialTimeout = cfg.DialTimeout
	opts.ReadTimeout = cfg.ReadTimeout
	opts.WriteTimeout = cfg.WriteTimeout

	client := goredis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, cfg.DialTimeout)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Client{Client: client}, nil
}

// Check reports whether Redis is reachable. It backs the readiness endpoint.
func (c *Client) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.Ping(ctx).Err()
}

// AsynqRedisOpt exposes the connection settings in the form the job queue
// expects, so both share one configured endpoint.
func AsynqRedisOpt(cfg config.RedisConfig) (*goredis.Options, error) {
	opts, err := goredis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse REDIS_URL: %w", err)
	}
	opts.DialTimeout = cfg.DialTimeout
	opts.ReadTimeout = cfg.ReadTimeout
	opts.WriteTimeout = cfg.WriteTimeout
	return opts, nil
}
