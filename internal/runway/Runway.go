package runway

import (
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"os"
)

type Runway struct {
	logger *zap.Logger
	redis  *redis.Client
}

func (r *Runway) Run() {

}

func Create() *Runway {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any

	return &Runway{
		logger: logger,
		redis:  ConnectRedis(),
	}
}

func ConnectRedis() *redis.Client {
	Addr := os.Getenv("REDIS_ADDR")
	Password := os.Getenv("REDIS_PASSWORD")

	cfg := &redis.Options{
		Addr: Addr,
		DB:   0,
	}

	// we may skip password in development
	if len(Password) > 0 {
		cfg.Password = Password
	}

	rdb := redis.NewClient(cfg)
	return rdb
}

func (r *Runway) Logger() *zap.Logger {
	return r.logger
}

func (r *Runway) Redis() *redis.Client {
	return r.redis
}
