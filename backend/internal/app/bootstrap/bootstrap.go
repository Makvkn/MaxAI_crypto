// Package bootstrap builds the dependency graph shared by the API and the
// worker. Both processes read the same configuration and talk to the same
// PostgreSQL and Redis instances; only their entrypoints differ (§204).
package bootstrap

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/maxaicrypto/backend/internal/app/config"
	appauth "github.com/maxaicrypto/backend/internal/application/auth"
	appai "github.com/maxaicrypto/backend/internal/application/ai"
	appportfolio "github.com/maxaicrypto/backend/internal/application/portfolio"
	apppricing "github.com/maxaicrypto/backend/internal/application/pricing"
	appscenarios "github.com/maxaicrypto/backend/internal/application/scenarios"
	appsync "github.com/maxaicrypto/backend/internal/application/sync"
	apptransactions "github.com/maxaicrypto/backend/internal/application/transactions"
	appusage "github.com/maxaicrypto/backend/internal/application/usage"
	appwallets "github.com/maxaicrypto/backend/internal/application/wallets"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	chainaddress "github.com/maxaicrypto/backend/internal/infrastructure/chain/address"
	googleauth "github.com/maxaicrypto/backend/internal/infrastructure/auth/google"
	jwtauth "github.com/maxaicrypto/backend/internal/infrastructure/auth/jwt"
	"github.com/maxaicrypto/backend/internal/infrastructure/auth/password"
	"github.com/maxaicrypto/backend/internal/infrastructure/observability"
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
)

// Version is stamped at build time and reported by the health endpoints.
var Version = "dev"

// Dependencies holds the long-lived resources of a process.
type Dependencies struct {
	Config    *config.Config
	Logger    *slog.Logger
	DB        *postgres.Pool
	Redis     *redis.Client
	Tx        *postgres.TxRunner
	Locker    *redis.Locker
	Providers *providers.Providers
	Jobs      *jobs.Client
	Auth      appauth.Service
	Tokens    appauth.TokenIssuer
	Wallets   appwallets.Service
	Portfolio    appportfolio.Service
	Performance  appportfolio.PerformanceService
	Transactions apptransactions.Service
	Usage        appusage.Service
	AI           appai.Service
	Scenarios    appscenarios.Service
	Sync         appsync.Service
	Pricing   apppricing.Service
	Snapshots appportfolio.SnapshotService
}

// New loads configuration, opens every connection and wires the providers.
// Anything misconfigured fails here rather than on the first request.
func New(ctx context.Context) (*Dependencies, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	logger := observability.NewLogger(os.Stdout, observability.LoggerOptions{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	}).With(slog.String("version", Version))

	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	redisClient, err := redis.NewClient(ctx, cfg.Redis)
	if err != nil {
		pool.Close()
		return nil, err
	}

	redisOpt, err := jobs.RedisOpt(cfg.Redis)
	if err != nil {
		_ = redisClient.Close()
		pool.Close()
		return nil, err
	}

	tx := postgres.NewTxRunner(pool)
	userRepo := userrepo.NewRepository(pool, tx)
	sessionRepo := userrepo.NewSessionRepository(pool, tx)
	subRepo := subscriptionrepo.NewRepository(pool)
	tokens := jwtauth.NewIssuer(cfg.Auth)
	authService := appauth.NewApp(
		userRepo,
		sessionRepo,
		subRepo,
		tokens,
		password.NewHasher(),
		googleauth.NewVerifier(cfg.Google),
		cfg,
	)
	jobsClient := jobs.NewClient(redisOpt, cfg.Sync)
	locker := redis.NewLocker(redisClient)
	providerBundle := providers.Build(cfg)

	assetRepo := assetrepo.NewRepository(pool, tx)
	positionRepo := positionrepo.NewRepository(pool, tx)
	transactionRepo := transactionrepo.NewRepository(pool, tx)
	priceRepo := pricerepo.NewRepository(pool, tx)
	snapshotRepo := portfoliorepo.NewSnapshotRepository(pool, tx)
	walletRepo := walletrepo.NewRepository(pool, tx)
	syncRepo := walletrepo.NewSyncRepository(pool)

	freshness := shared.FreshnessThresholds{
		FreshMax:  cfg.Cache.FreshnessFreshMax,
		RecentMax: cfg.Cache.FreshnessRecentMax,
		StaleMax:  cfg.Cache.FreshnessStaleMax,
	}
	pricingService := apppricing.NewApp(assetRepo, priceRepo, providerBundle.Resolver, !cfg.Provider.HasBlockchainCredentials())
	calculator := appportfolio.NewCalculator(assetRepo, positionRepo, pricingService, freshness)
	snapshotService := appportfolio.NewSnapshotApp(calculator, snapshotRepo)
	portfolioService := appportfolio.NewReadApp(walletRepo, syncRepo, calculator)
	performanceService := appportfolio.NewPerformanceApp(walletRepo, syncRepo, snapshotRepo, assetRepo)
	transactionService := apptransactions.NewApp(walletRepo, syncRepo, transactionRepo, assetRepo, pricingService)
	entitlements := appusage.NewEntitlementApp(subRepo, walletRepo, cfg.AI.DailyLimit)
	usageRepo := usagerepo.NewRepository(pool, tx)
	usageCounter := redis.NewUsageCounter(redisClient)
	usageService := appusage.NewApp(usageRepo, usageCounter, entitlements)
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

	return &Dependencies{
		Config:    cfg,
		Logger:    logger,
		DB:        pool,
		Redis:     redisClient,
		Tx:        tx,
		Locker:    locker,
		Providers: providerBundle,
		Jobs:      jobsClient,
		Auth:      authService,
		Tokens:    tokens,
		Wallets:   walletService,
		Portfolio:    portfolioService,
		Performance:  performanceService,
		Transactions: transactionService,
		Usage:        usageService,
		AI:           aiService,
		Scenarios:    scenarioService,
		Sync:         syncService,
		Pricing:   pricingService,
		Snapshots: snapshotService,
	}, nil
}

// Close releases every resource in reverse order of acquisition.
func (d *Dependencies) Close() error {
	var errs []error
	if d.Jobs != nil {
		errs = append(errs, d.Jobs.Close())
	}
	if d.Redis != nil {
		errs = append(errs, d.Redis.Close())
	}
	if d.DB != nil {
		d.DB.Close()
	}
	return errors.Join(errs...)
}
