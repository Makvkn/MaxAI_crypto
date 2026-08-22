package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/hibiken/asynq"

	"github.com/maxaicrypto/backend/internal/app/config"
	"github.com/maxaicrypto/backend/internal/infrastructure/observability"
)

// Server runs the background worker: the task processor plus the periodic
// scheduler that enqueues recurring work (§58, §62).
type Server struct {
	server    *asynq.Server
	scheduler *asynq.Scheduler
	mux       *asynq.ServeMux
	logger    *slog.Logger
}

// NewServer builds the worker. Queue weights keep user-visible syncs ahead of
// periodic maintenance.
func NewServer(opt asynq.RedisConnOpt, cfg *config.Config, logger *slog.Logger) *Server {
	server := asynq.NewServer(opt, asynq.Config{
		Concurrency: cfg.Worker.Concurrency,
		Queues: map[string]int{
			QueueCritical: 6,
			QueueDefault:  3,
			QueueLow:      1,
		},
		ShutdownTimeout: cfg.Worker.ShutdownTimeout,
		Logger:          &asynqLogger{logger: logger},
		// RetryDelayFunc backs off exponentially so a struggling provider is
		// not hammered by retries (§158).
		RetryDelayFunc: asynq.DefaultRetryDelayFunc,
		ErrorHandler:   asynq.ErrorHandlerFunc(newErrorHandler(logger)),
	})

	scheduler := asynq.NewScheduler(opt, &asynq.SchedulerOpts{
		Location: time.UTC,
		Logger:   &asynqLogger{logger: logger},
	})

	return &Server{
		server:    server,
		scheduler: scheduler,
		mux:       asynq.NewServeMux(),
		logger:    logger,
	}
}

// Mux exposes the handler multiplexer so handlers can register themselves.
func (s *Server) Mux() *asynq.ServeMux { return s.mux }

// RegisterPeriodic schedules the recurring jobs: wallet synchronization at the
// configured interval and a price refresh aligned with the price cache TTL
// (§62, §120).
func (s *Server) RegisterPeriodic(cfg *config.Config) error {
	if _, err := s.scheduler.Register(
		cronEvery(cfg.Sync.Interval),
		asynq.NewTask(TypeSyncScheduler, nil),
		asynq.Queue(QueueLow),
	); err != nil {
		return err
	}

	priceRefresh, err := NewPriceRefresh(nil)
	if err != nil {
		return err
	}
	if _, err := s.scheduler.Register(
		cronEvery(cfg.Cache.PriceTTL),
		priceRefresh,
		asynq.Queue(QueueLow),
	); err != nil {
		return err
	}
	return nil
}

// Run starts the processor and the scheduler and blocks until ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	if err := s.scheduler.Start(); err != nil {
		return err
	}
	defer s.scheduler.Shutdown()

	if err := s.server.Start(s.mux); err != nil {
		return err
	}

	<-ctx.Done()
	s.logger.Info("worker shutting down")
	s.server.Shutdown()
	return nil
}

// cronEvery renders an interval as an Asynq "@every" spec.
func cronEvery(d time.Duration) string {
	return "@every " + d.String()
}

// newErrorHandler logs a failing task once its retries are exhausted, which is
// the dead-letter signal operators act on (§59).
func newErrorHandler(logger *slog.Logger) func(context.Context, *asynq.Task, error) {
	return func(ctx context.Context, task *asynq.Task, err error) {
		retried, _ := asynq.GetRetryCount(ctx)
		maxRetry, _ := asynq.GetMaxRetry(ctx)
		jobID, _ := asynq.GetTaskID(ctx)

		attrs := []any{
			slog.String(observability.FieldJobID, jobID),
			slog.String("task_type", task.Type()),
			slog.Int("attempt", retried+1),
			slog.Any(observability.FieldError, err),
		}
		if retried >= maxRetry {
			logger.ErrorContext(ctx, "task exhausted retries", attrs...)
			return
		}
		logger.WarnContext(ctx, "task failed, will retry", attrs...)
	}
}

// asynqLogger adapts slog to the asynq.Logger interface so worker output shares
// the application's structured format (§122).
type asynqLogger struct {
	logger *slog.Logger
}

func (l *asynqLogger) Debug(args ...any) { l.logger.Debug(concat(args)) }
func (l *asynqLogger) Info(args ...any)  { l.logger.Info(concat(args)) }
func (l *asynqLogger) Warn(args ...any)  { l.logger.Warn(concat(args)) }
func (l *asynqLogger) Error(args ...any) { l.logger.Error(concat(args)) }
func (l *asynqLogger) Fatal(args ...any) { l.logger.Error(concat(args)) }

func concat(args []any) string { return strings.TrimSpace(fmt.Sprintln(args...)) }
