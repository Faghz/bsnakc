package bootstrap

import (
	"context"

	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/config"
	"github.com/elzestia/go-boilerplate/internal/shared/database"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func initDatabase(cfg *config.Config, logger ports.Logger) (*sqlx.DB, error) {
	logger.Info(context.Background(), "Connecting to PostgreSQL...")
	db, err := database.NewPostgresDB(&cfg.Database, &cfg.OTel)
	if err != nil {
		return nil, err
	}
	logger.Info(context.Background(), "PostgreSQL connected")
	return db, nil
}

func initRedis(cfg *config.Config, logger ports.Logger) (*redis.Client, error) {
	logger.Info(context.Background(), "Connecting to Redis...")
	client, err := database.NewRedisClient(&cfg.Redis, &cfg.OTel)
	if err != nil {
		return nil, err
	}
	logger.Info(context.Background(), "Redis connected")
	return client, nil
}
