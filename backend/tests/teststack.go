//go:build integration

package tests

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/hibiken/asynq"

	appauth "github.com/maxaicrypto/backend/internal/application/auth"
	appai "github.com/maxaicrypto/backend/internal/application/ai"
	appportfolio "github.com/maxaicrypto/backend/internal/application/portfolio"
	apppricing "github.com/maxaicrypto/backend/internal/application/pricing"
	appscenarios "github.com/maxaicrypto/backend/internal/application/scenarios"
	appsync "github.com/maxaicrypto/backend/internal/application/sync"
	apptransactions "github.com/maxaicrypto/backend/internal/application/transactions"
	appusage "github.com/maxaicrypto/backend/internal/application/usage"
	appwallets "github.com/maxaicrypto/backend/internal/application/wallets"
	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	chainaddress "github.com/maxaicrypto/backend/internal/infrastructure/chain/address"
	googleauth "github.com/maxaicrypto/backend/internal/infrastructure/auth/google"
	jwtauth "github.com/maxaicrypto/backend/internal/infrastructure/auth/jwt"
	"github.com/maxaicrypto/backend/internal/infrastructure/auth/password"
	"github.com/maxaicrypto/backend/internal/infrastructure/postgres"
	assetrepo "github.com/maxaicrypto/backend/internal/infrastructure/postgres/asset"
	chainrepo "github.com/maxaicrypto/backend/internal/infrastructure/postgres/chain"
	conversationrepo "github.com/maxaicrypto/backend/internal/infrastructure/postgres/conversation"
	portfoliorepo "github.com/maxaicrypto/backend/internal/infrastructure/postgres/portfolio"
	scenariorepo "github.com/maxaicrypto/backend/internal/infrastructure/postgres/scenario"
	usagerepo "github.com/maxaicrypto/backend/internal/infrastructure/postgres/usage"
	positionrepo "github.com/maxaicrypto/backend/internal/infrastructure/postgres/position"
	pricerepo "github.com/maxaicrypto/backend/internal/infrastructure/postgres/price"
	subscriptionrepo "github.com/maxaicrypto/backend/internal/infrastructure/postgres/subscription"
	transactionrepo "github.com/maxaicrypto/backend/internal/infrastructure/postgres/transaction"
	userrepo "github.com/maxaicrypto/backend/internal/infrastructure/postgres/user"
	walletrepo "github.com/maxaicrypto/backend/internal/infrastructure/postgres/wallet"
	"github.com/maxaicrypto/backend/internal/infrastructure/redis"
	"github.com/maxaicrypto/backend/internal/jobs"
	"github.com/maxaicrypto/backend/internal/providers"
	transport "github.com/maxaicrypto/backend/internal/transport/http"
	"github.com/maxaicrypto/backend/internal/transport/http/handlers"
	"github.com/maxaicrypto/backend/tests/testsupport"
)

type testStack struct {
	Config    *config.Config
	Sync      appsync.Service
	Portfolio    appportfolio.Service
	Performance  appportfolio.PerformanceService
	Transactions apptransactions.Service
	Usage        appusage.Service
	AI           appai.Service
	Scenarios    appscenarios.Service
	Pricing      apppricing.Service
	Snapshots appportfolio.SnapshotService
}

func newWalletsTestStack(t *testing.T, pool *postgres.Pool) (http.Handler, *config.Config, testStack) {
	t.Helper()

	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		redisURL = os.Getenv("REDIS_URL")
	}
	if redisURL == "" {
		t.Skip("TEST_REDIS_URL is not set")
	}

	cfg := &config.Config{
		Env: config.EnvDevelopment,
		HTTP: config.HTTPConfig{
			MaxRequestBytes: 1 << 20,
			AllowedOrigins:  []string{"http://localhost:5173"},
		},
		Auth: config.AuthConfig{
			JWTSecret:       "change-me-development-secret-change-me-please",
			JWTIssuer:       "maxai-crypto",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 720 * time.Hour,
		},
		AI:    config.AIConfig{DailyLimit: 10},
		Redis: config.RedisConfig{
			URL:          redisURL,
			DialTimeout:  10 * time.Second,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		Cache: config.CacheConfig{
			FreshnessFreshMax:  5 * time.Minute,
			FreshnessRecentMax: 15 * time.Minute,
			FreshnessStaleMax:  60 * time.Minute,
		},
		Sync: config.SyncConfig{
			Interval:   15 * time.Minute,
			LockTTL:    2 * time.Minute,
			MaxRetries: 3,
			JobTimeout: 5 * time.Minute,
		},
	}

	tx := postgres.NewTxRunner(pool)
	tokens := jwtauth.NewIssuer(cfg.Auth)
	subRepo := subscriptionrepo.NewRepository(pool)
	walletRepo := walletrepo.NewRepository(pool, tx)
	syncRepo := walletrepo.NewSyncRepository(pool)
	jobsClient := jobs.NewClient(mustRedisOpt(t, cfg.Redis), cfg.Sync)
	t.Cleanup(func() { _ = jobsClient.Close() })

	redisClient := testsupport.Redis(t)
	locker := redis.NewLocker(redisClient)
	providerBundle := providers.Build(cfg)

	assetRepo := assetrepo.NewRepository(pool, tx)
	positionRepo := positionrepo.NewRepository(pool, tx)
	transactionRepo := transactionrepo.NewRepository(pool, tx)
	priceRepo := pricerepo.NewRepository(pool, tx)
	snapshotRepo := portfoliorepo.NewSnapshotRepository(pool, tx)
	freshness := shared.FreshnessThresholds{
		FreshMax:  cfg.Cache.FreshnessFreshMax,
		RecentMax: cfg.Cache.FreshnessRecentMax,
		StaleMax:  cfg.Cache.FreshnessStaleMax,
	}
	pricingService := apppricing.NewApp(assetRepo, priceRepo, providerBundle.Resolver, true)
	calculator := appportfolio.NewCalculator(assetRepo, positionRepo, pricingService, freshness)
	snapshotService := appportfolio.NewSnapshotApp(calculator, snapshotRepo)
	portfolioService := appportfolio.NewReadApp(walletRepo, syncRepo, calculator)
	performanceService := appportfolio.NewPerformanceApp(walletRepo, syncRepo, snapshotRepo, assetRepo)
	transactionService := apptransactions.NewApp(walletRepo, syncRepo, transactionRepo, assetRepo, pricingService)
	entitlements := appusage.NewEntitlementApp(subRepo, walletRepo, cfg.AI.DailyLimit)
	usageService := appusage.NewApp(usagerepo.NewRepository(pool, tx), redis.NewUsageCounter(redisClient), entitlements)
	conversationRepo := conversationrepo.NewRepository(pool, tx)
	scenarioService := appscenarios.NewApp(
		portfolioService,
		assetRepo,
		scenariorepo.NewRepository(pool),
		usageService,
		entitlements,
		providerBundle.Resolver,
		cfg.AI,
	)
	orchestrator := appai.NewOrchestrator(appai.OrchestratorDeps{
		Portfolios:   portfolioService,
		Performance:  performanceService,
		Transactions: transactionService,
		Pricing:      pricingService,
		Scenarios:    scenarioService,
		Snapshots:    snapshotRepo,
		Resolver:     providerBundle.Resolver,
		AI:           cfg.AI,
	})
	aiService := appai.NewApp(conversationRepo, walletRepo, syncRepo, usageService, orchestrator)
	syncService := appsync.NewApp(
		walletRepo,
		syncRepo,
		positionRepo,
		transactionRepo,
		assetRepo,
		snapshotService,
		pricingService,
		nil,
		providerBundle.Resolver,
		locker,
		cfg.Sync,
	)

	authService := appauth.NewApp(
		userrepo.NewRepository(pool, tx),
		userrepo.NewSessionRepository(pool, tx),
		subRepo,
		tokens,
		password.NewHasher(),
		googleauth.NewVerifier(cfg.Google),
		cfg,
	)
	walletService := appwallets.NewApp(
		walletRepo,
		syncRepo,
		chainrepo.NewRepository(pool),
		chainaddress.NewValidator(),
		entitlements,
		jobsClient,
		tx,
		freshness,
	)

	router := transport.NewRouter(transport.RouterDeps{
		Config:  cfg,
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		Health:  handlers.NewHealthHandler("test", nil),
		Auth:    handlers.NewAuthHandler(authService),
		Wallets: handlers.NewWalletsHandler(walletService),
		Portfolio:    handlers.NewPortfolioHandler(portfolioService, freshness),
		Performance:  handlers.NewPerformanceHandler(performanceService),
		Transactions: handlers.NewTransactionsHandler(transactionService),
		AI:           handlers.NewAIHandler(aiService, usageService, scenarioService),
		Tokens:       tokens,
		RateLimiter:  redis.NewRateLimiter(redisClient),
	})
	return router, cfg, testStack{
		Config:       cfg,
		Sync:         syncService,
		Portfolio:    portfolioService,
		Performance:  performanceService,
		Transactions: transactionService,
		Usage:        usageService,
		AI:           aiService,
		Scenarios:    scenarioService,
		Pricing:      pricingService,
		Snapshots:    snapshotService,
	}
}

func guestAccessToken(t *testing.T, router http.Handler) string {
	t.Helper()
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/auth/guest", nil))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create guest status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var session struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &session); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	return session.AccessToken
}

func mustRedisOpt(t *testing.T, cfg config.RedisConfig) asynq.RedisClientOpt {
	t.Helper()
	opt, err := jobs.RedisOpt(cfg)
	if err != nil {
		t.Fatalf("redis opt: %v", err)
	}
	return opt
}

func newWalletsTestRouter(t *testing.T, pool *postgres.Pool) (http.Handler, *config.Config) {
	router, cfg, _ := newWalletsTestStack(t, pool)
	return router, cfg
}
