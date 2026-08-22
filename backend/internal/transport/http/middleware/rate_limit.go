package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/domain/apperr"
	"github.com/maxaicrypto/backend/internal/infrastructure/redis"
	"github.com/maxaicrypto/backend/internal/transport/http/response"
)

// RateLimit applies the IP / anonymous / authenticated layers from §153.
// OptionalAuthenticate should run before this middleware so authenticated
// subjects can be identified without rejecting public routes.
func RateLimit(limiter *redis.RateLimiter, cfg config.RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limiter == nil {
				next.ServeHTTP(w, r)
				return
			}

			ip := clientIP(r)
			ok, err := limiter.Allow(r.Context(), "ip", ip, cfg.IPPerMinute)
			if err != nil {
				response.Error(w, r, apperr.Wrap(apperr.CodeInternal, err))
				return
			}
			if !ok {
				response.Error(w, r, apperr.New(apperr.CodeRateLimit))
				return
			}

			if userID, authed := UserIDFrom(r.Context()); authed {
				ok, err = limiter.Allow(r.Context(), "user", userID.String(), cfg.AuthenticatedPerMinute)
			} else {
				ok, err = limiter.Allow(r.Context(), "anon", ip, cfg.AnonymousPerMinute)
			}
			if err != nil {
				response.Error(w, r, apperr.Wrap(apperr.CodeInternal, err))
				return
			}
			if !ok {
				response.Error(w, r, apperr.New(apperr.CodeRateLimit))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
