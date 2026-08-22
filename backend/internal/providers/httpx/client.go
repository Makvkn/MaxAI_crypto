// Package httpx holds the HTTP plumbing shared by provider adapters: timeouts,
// bounded retries with exponential backoff, and the translation of transport
// failures into apperr codes (§27, §28, §158).
package httpx

import (
	"context"
	"errors"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/domain/provider"
)

// Config configures one provider's HTTP behaviour.
type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
	// MaxAttempts includes the first attempt.
	MaxAttempts int
	// BackoffSchedule is the delay before attempt N+1. When it is shorter than
	// MaxAttempts the last entry repeats.
	BackoffSchedule []time.Duration
}

// Client is a provider-scoped HTTP client.
type Client struct {
	provider provider.Name
	cfg      Config
	http     *http.Client
}

// New builds a client for a provider.
func New(name provider.Name, cfg Config) *Client {
	return &Client{
		provider: name,
		cfg:      cfg,
		http:     &http.Client{Timeout: cfg.Timeout},
	}
}

// BaseURL returns the configured base URL.
func (c *Client) BaseURL() string { return c.cfg.BaseURL }

// APIKey returns the configured credential. Adapters attach it in the header
// form their vendor expects; it is never logged (§124).
func (c *Client) APIKey() string { return c.cfg.APIKey }

// Do executes req, retrying transport failures and retryable status codes up to
// MaxAttempts. Callers own the response body.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	attempts := max(c.cfg.MaxAttempts, 1)

	var lastErr error
	for attempt := range attempts {
		if attempt > 0 {
			if err := c.wait(ctx, attempt-1); err != nil {
				return nil, err
			}
		}

		resp, err := c.http.Do(req.Clone(ctx))
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, apperr.Wrap(apperr.CodeProviderError, err).
					WithDetail("provider", string(c.provider))
			}
			lastErr = err
			continue
		}
		if !isRetryable(resp.StatusCode) {
			return resp, nil
		}
		lastErr = errors.New(resp.Status)
		_ = resp.Body.Close()
	}

	return nil, apperr.Wrap(apperr.CodeProviderError, lastErr).
		WithDetail("provider", string(c.provider)).
		WithDetail("attempts", attempts)
}

// wait sleeps for the backoff of the given retry index, with jitter so that
// concurrent syncs do not retry in lockstep.
func (c *Client) wait(ctx context.Context, index int) error {
	delay := c.backoff(index)
	if delay <= 0 {
		return nil
	}
	jittered := delay/2 + time.Duration(rand.Int64N(int64(delay)))
	timer := time.NewTimer(jittered)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return apperr.Wrap(apperr.CodeProviderError, ctx.Err()).
			WithDetail("provider", string(c.provider))
	case <-timer.C:
		return nil
	}
}

func (c *Client) backoff(index int) time.Duration {
	schedule := c.cfg.BackoffSchedule
	if len(schedule) == 0 {
		return 0
	}
	if index >= len(schedule) {
		index = len(schedule) - 1
	}
	return schedule[index]
}

// isRetryable reports whether a status is worth another attempt. Client errors
// other than rate limiting are permanent and retrying them only wastes budget.
func isRetryable(status int) bool {
	switch status {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// MapStatus translates a provider response status into a domain error. Provider
// error shapes never reach business logic (§28).
func MapStatus(name provider.Name, status int) error {
	switch {
	case status < 400:
		return nil
	case status == http.StatusTooManyRequests:
		return apperr.New(apperr.CodeRateLimit).WithDetail("provider", string(name))
	case status == http.StatusNotFound:
		return apperr.New(apperr.CodeNotFound).WithDetail("provider", string(name))
	default:
		return apperr.New(apperr.CodeProviderError).
			WithDetail("provider", string(name)).
			WithDetail("status", status)
	}
}
