package jobs

import (
	"github.com/hibiken/asynq"

	"github.com/maxaicrypto/backend/internal/app/config"
	appredis "github.com/maxaicrypto/backend/internal/infrastructure/redis"
)

// RedisOpt translates the shared Redis configuration into the connection
// options Asynq expects, so the API and the worker address one endpoint (§64).
func RedisOpt(cfg config.RedisConfig) (asynq.RedisClientOpt, error) {
	opts, err := appredis.AsynqRedisOpt(cfg)
	if err != nil {
		return asynq.RedisClientOpt{}, err
	}
	return asynq.RedisClientOpt{
		Addr:         opts.Addr,
		Username:     opts.Username,
		Password:     opts.Password,
		DB:           opts.DB,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		TLSConfig:    opts.TLSConfig,
	}, nil
}
