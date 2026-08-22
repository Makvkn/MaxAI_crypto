// Package config turns environment variables into a validated, typed
// configuration tree. Every threshold, TTL, limit and backoff schedule the
// backend relies on is loaded here rather than hardcoded at the point of use,
// as required by the backend specification (§65, §135).
package config

import (
	"time"
)

// Environment identifies the deployment the process is running in.
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
)

func (e Environment) IsProduction() bool { return e == EnvProduction }

// Config is the fully resolved backend configuration.
type Config struct {
	Env      Environment
	HTTP     HTTPConfig
	Log      LogConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Google   GoogleConfig
	Provider ProviderConfig
	AI       AIConfig
	Sync     SyncConfig
	Cache    CacheConfig
	Limits   RateLimitConfig
	Worker   WorkerConfig
}

// HTTPConfig covers the public API listener.
type HTTPConfig struct {
	Port              int
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxRequestBytes   int64
	AllowedOrigins    []string
}

// LogConfig controls the structured logger.
type LogConfig struct {
	Level  string
	Format string
}

// DatabaseConfig describes the PostgreSQL connection pool.
type DatabaseConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	ConnectTimeout  time.Duration
}

// RedisConfig describes the Redis connection used for cache, locks, rate
// limiting and the job queue (§64).
type RedisConfig struct {
	URL          string
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// AuthConfig holds token issuance settings. Refresh sessions are tracked
// server-side, so only the access-token signing material lives here (§13).
type AuthConfig struct {
	JWTSecret       string
	JWTIssuer       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// GoogleConfig holds OAuth/OIDC credentials (§15).
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// ProviderConfig holds shared provider transport settings plus per-provider
// credentials. Credentials never leave the backend (§151).
type ProviderConfig struct {
	Timeout         time.Duration
	MaxAttempts     int
	BackoffSchedule []time.Duration

	Zerion    ProviderCredentials
	Tatum     ProviderCredentials
	CoinGecko ProviderCredentials
}

// ProviderCredentials is the API key and base URL of one external provider.
type ProviderCredentials struct {
	APIKey  string
	BaseURL string
}

// HasBlockchainCredentials reports whether any blockchain provider is configured.
func (c ProviderConfig) HasBlockchainCredentials() bool {
	return c.Zerion.APIKey != "" || c.Tatum.APIKey != ""
}

// HasAICredentials reports whether an LLM provider is configured.
func (c AIConfig) HasAICredentials() bool {
	return c.APIKey != ""
}

// AIConfig holds LLM selection and the cost controls from §176 and §177.
// The model is configuration and must never be hardcoded in domain logic.
type AIConfig struct {
	Provider           string
	Model              string
	APIKey             string
	BaseURL            string
	RequestTimeout     time.Duration
	MaxOutputTokens    int
	MaxToolCalls       int
	MaxContextMessages int
	DailyLimit         int
}

// SyncConfig holds wallet synchronization scheduling and locking (§61, §62).
type SyncConfig struct {
	Interval   time.Duration
	LockTTL    time.Duration
	JobTimeout time.Duration
	MaxRetries int
}

// CacheConfig holds Redis TTLs and the price freshness thresholds. Cache TTL
// and freshness are related but distinct concepts (§37, §120).
type CacheConfig struct {
	PriceTTL     time.Duration
	PortfolioTTL time.Duration

	FreshnessFreshMax  time.Duration
	FreshnessRecentMax time.Duration
	FreshnessStaleMax  time.Duration
}

// RateLimitConfig holds the separate limiter layers from §153.
type RateLimitConfig struct {
	IPPerMinute            int
	AnonymousPerMinute     int
	AuthenticatedPerMinute int
}

// WorkerConfig holds background worker settings.
type WorkerConfig struct {
	Concurrency int
	// ShutdownTimeout bounds how long a running job may finish after a stop
	// signal before the worker exits and the job is retried elsewhere.
	ShutdownTimeout time.Duration
}

// Load reads configuration from the environment and validates it. All problems
// are reported together.
func Load() (*Config, error) {
	l := &loader{}

	env := Environment(l.str("APP_ENV", string(EnvDevelopment)))
	switch env {
	case EnvDevelopment, EnvStaging, EnvProduction:
	default:
		l.fail("APP_ENV", "must be one of development, staging, production")
	}

	cfg := &Config{
		Env: env,
		HTTP: HTTPConfig{
			Port:              l.int("HTTP_PORT", 8080),
			ReadHeaderTimeout: l.duration("HTTP_READ_HEADER_TIMEOUT", 10*time.Second),
			WriteTimeout:      l.duration("HTTP_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:       l.duration("HTTP_IDLE_TIMEOUT", 120*time.Second),
			ShutdownTimeout:   l.duration("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second),
			MaxRequestBytes:   l.int64("HTTP_MAX_REQUEST_BYTES", 1<<20),
			AllowedOrigins:    l.list("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
		},
		Log: LogConfig{
			Level:  l.str("LOG_LEVEL", "info"),
			Format: l.str("LOG_FORMAT", "json"),
		},
		Database: DatabaseConfig{
			URL:             l.required("DATABASE_URL"),
			MaxConns:        int32(l.int("DATABASE_MAX_CONNS", 20)),
			MinConns:        int32(l.int("DATABASE_MIN_CONNS", 2)),
			MaxConnLifetime: l.duration("DATABASE_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime: l.duration("DATABASE_MAX_CONN_IDLE_TIME", 30*time.Minute),
			ConnectTimeout:  l.duration("DATABASE_CONNECT_TIMEOUT", 10*time.Second),
		},
		Redis: RedisConfig{
			URL:          l.required("REDIS_URL"),
			DialTimeout:  l.duration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  l.duration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: l.duration("REDIS_WRITE_TIMEOUT", 3*time.Second),
		},
		Auth: AuthConfig{
			JWTSecret:       l.required("AUTH_JWT_SECRET"),
			JWTIssuer:       l.str("AUTH_JWT_ISSUER", "maxai-crypto"),
			AccessTokenTTL:  l.duration("AUTH_ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL: l.duration("AUTH_REFRESH_TOKEN_TTL", 720*time.Hour),
		},
		Google: GoogleConfig{
			ClientID:     l.str("GOOGLE_CLIENT_ID", ""),
			ClientSecret: l.str("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  l.str("GOOGLE_REDIRECT_URL", ""),
		},
		Provider: ProviderConfig{
			Timeout:     l.duration("PROVIDER_TIMEOUT", 15*time.Second),
			MaxAttempts: l.int("PROVIDER_MAX_ATTEMPTS", 3),
			BackoffSchedule: l.durations("PROVIDER_BACKOFF_SCHEDULE", []time.Duration{
				30 * time.Second, time.Minute, 5 * time.Minute,
			}),
			Zerion: ProviderCredentials{
				APIKey:  l.str("ZERION_API_KEY", ""),
				BaseURL: l.str("ZERION_BASE_URL", "https://api.zerion.io"),
			},
			Tatum: ProviderCredentials{
				APIKey:  l.str("TATUM_API_KEY", ""),
				BaseURL: l.str("TATUM_BASE_URL", "https://api.tatum.io"),
			},
			CoinGecko: ProviderCredentials{
				APIKey:  l.str("COINGECKO_API_KEY", ""),
				BaseURL: l.str("COINGECKO_BASE_URL", "https://api.coingecko.com/api/v3"),
			},
		},
		AI: AIConfig{
			Provider:           l.str("LLM_PROVIDER", "openai"),
			Model:              l.str("LLM_MODEL", "gpt-4o-mini"),
			APIKey:             l.str("OPENAI_API_KEY", ""),
			BaseURL:            l.str("OPENAI_BASE_URL", "https://api.openai.com/v1"),
			RequestTimeout:     l.duration("LLM_REQUEST_TIMEOUT", 60*time.Second),
			MaxOutputTokens:    l.int("LLM_MAX_OUTPUT_TOKENS", 1500),
			MaxToolCalls:       l.int("AI_MAX_TOOL_CALLS", 5),
			MaxContextMessages: l.int("AI_MAX_CONTEXT_MESSAGES", 10),
			DailyLimit:         l.int("AI_DAILY_LIMIT", 10),
		},
		Sync: SyncConfig{
			Interval:   l.duration("SYNC_INTERVAL", 15*time.Minute),
			LockTTL:    l.duration("SYNC_LOCK_TTL", 10*time.Minute),
			JobTimeout: l.duration("SYNC_JOB_TIMEOUT", 10*time.Minute),
			MaxRetries: l.int("SYNC_JOB_MAX_RETRIES", 3),
		},
		Cache: CacheConfig{
			PriceTTL:           l.duration("PRICE_CACHE_TTL", 45*time.Second),
			PortfolioTTL:       l.duration("PORTFOLIO_CACHE_TTL", 2*time.Minute),
			FreshnessFreshMax:  l.duration("FRESHNESS_FRESH_MAX", 5*time.Minute),
			FreshnessRecentMax: l.duration("FRESHNESS_RECENT_MAX", 15*time.Minute),
			FreshnessStaleMax:  l.duration("FRESHNESS_STALE_MAX", 60*time.Minute),
		},
		Limits: RateLimitConfig{
			IPPerMinute:            l.int("RATE_LIMIT_IP_PER_MINUTE", 120),
			AnonymousPerMinute:     l.int("RATE_LIMIT_ANONYMOUS_PER_MINUTE", 60),
			AuthenticatedPerMinute: l.int("RATE_LIMIT_AUTHENTICATED_PER_MINUTE", 240),
		},
		Worker: WorkerConfig{
			Concurrency:     l.int("WORKER_CONCURRENCY", 10),
			ShutdownTimeout: l.duration("WORKER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
	}

	cfg.validate(l)
	if err := l.err(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate(l *loader) {
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		l.fail("HTTP_PORT", "must be between 1 and 65535")
	}
	if c.Database.MinConns > c.Database.MaxConns {
		l.fail("DATABASE_MIN_CONNS", "must not exceed DATABASE_MAX_CONNS")
	}
	if len(c.Auth.JWTSecret) < 32 {
		l.fail("AUTH_JWT_SECRET", "must be at least 32 characters")
	}
	if c.Auth.AccessTokenTTL >= c.Auth.RefreshTokenTTL {
		l.fail("AUTH_ACCESS_TOKEN_TTL", "must be shorter than AUTH_REFRESH_TOKEN_TTL")
	}
	if c.AI.DailyLimit < 0 {
		l.fail("AI_DAILY_LIMIT", "must not be negative")
	}
	if c.AI.MaxToolCalls < 1 {
		l.fail("AI_MAX_TOOL_CALLS", "must be at least 1")
	}

	// Freshness buckets must be strictly increasing or the classification in
	// §37 becomes ambiguous.
	if !(c.Cache.FreshnessFreshMax < c.Cache.FreshnessRecentMax &&
		c.Cache.FreshnessRecentMax < c.Cache.FreshnessStaleMax) {
		l.fail("FRESHNESS_FRESH_MAX", "thresholds must strictly increase: fresh < recent < stale")
	}

	if len(c.Provider.BackoffSchedule) == 0 {
		l.fail("PROVIDER_BACKOFF_SCHEDULE", "must contain at least one duration")
	}

	if c.Env.IsProduction() {
		for key, value := range map[string]string{
			"ZERION_API_KEY":    c.Provider.Zerion.APIKey,
			"TATUM_API_KEY":     c.Provider.Tatum.APIKey,
			"COINGECKO_API_KEY": c.Provider.CoinGecko.APIKey,
			"OPENAI_API_KEY":    c.AI.APIKey,
			"GOOGLE_CLIENT_ID":  c.Google.ClientID,
		} {
			if value == "" {
				l.fail(key, "is required in production")
			}
		}
		for _, origin := range c.HTTP.AllowedOrigins {
			if origin == "*" {
				l.fail("CORS_ALLOWED_ORIGINS", "must not be a wildcard in production")
			}
		}
	}
}
