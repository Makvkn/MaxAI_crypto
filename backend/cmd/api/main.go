// Command api serves the REST API. Request handling and background processing
// run as separate processes so a slow sync never blocks a user request (§204).
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/maxaicrypto/backend/internal/app/bootstrap"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/infrastructure/redis"
	transport "github.com/maxaicrypto/backend/internal/transport/http"
	"github.com/maxaicrypto/backend/internal/transport/http/handlers"
)

func main() {
	if err := run(); err != nil {
		// The logger may not exist yet when configuration fails, so startup
		// errors go to stderr directly.
		fmt.Fprintf(os.Stderr, "api startup failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	deps, err := bootstrap.New(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := deps.Close(); closeErr != nil {
			deps.Logger.Error("shutdown cleanup failed", slog.Any("error", closeErr))
		}
	}()

	router := transport.NewRouter(transport.RouterDeps{
		Config: deps.Config,
		Logger: deps.Logger,
		Health: handlers.NewHealthHandler(bootstrap.Version, map[string]handlers.Checker{
			"postgres": deps.DB,
			"redis":    deps.Redis,
		}),
		Auth:      handlers.NewAuthHandler(deps.Auth),
		Wallets:   handlers.NewWalletsHandler(deps.Wallets),
		Portfolio: handlers.NewPortfolioHandler(deps.Portfolio, shared.FreshnessThresholds{
			FreshMax:  deps.Config.Cache.FreshnessFreshMax,
			RecentMax: deps.Config.Cache.FreshnessRecentMax,
			StaleMax:  deps.Config.Cache.FreshnessStaleMax,
		}),
		Performance:  handlers.NewPerformanceHandler(deps.Performance),
		Transactions: handlers.NewTransactionsHandler(deps.Transactions),
		AI:           handlers.NewAIHandler(deps.AI, deps.Usage, deps.Scenarios),
		Tokens:       deps.Tokens,
		RateLimiter:  redis.NewRateLimiter(deps.Redis),
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", deps.Config.HTTP.Port),
		Handler:           router,
		ReadHeaderTimeout: deps.Config.HTTP.ReadHeaderTimeout,
		// WriteTimeout stays unset: SSE responses are long-lived by design and
		// a global write deadline would cut AI streams short (§82).
		IdleTimeout: deps.Config.HTTP.IdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		deps.Logger.Info("api listening", slog.Int("port", deps.Config.HTTP.Port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		deps.Logger.Info("api shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), deps.Config.HTTP.ShutdownTimeout)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}
