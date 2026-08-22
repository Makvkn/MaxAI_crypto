// Command worker runs background jobs: wallet synchronization, portfolio
// snapshots and price refresh (§58). It shares configuration and connections
// with the API but scales independently (§204).
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/maxaicrypto/backend/internal/app/bootstrap"
	"github.com/maxaicrypto/backend/internal/jobs"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "worker startup failed: %v\n", err)
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

	redisOpt, err := jobs.RedisOpt(deps.Config.Redis)
	if err != nil {
		return err
	}

	server := jobs.NewServer(redisOpt, deps.Config, deps.Logger)
	// Application services are attached here as each vertical slice lands;
	// until then a task fails loudly rather than silently succeeding.
	jobs.NewHandlers(deps.Sync, deps.Snapshots, deps.Pricing, deps.Jobs, deps.Logger).Register(server.Mux())

	if err := server.RegisterPeriodic(deps.Config); err != nil {
		return err
	}

	deps.Logger.Info("worker started", slog.Int("concurrency", deps.Config.Worker.Concurrency))
	return server.Run(ctx)
}
