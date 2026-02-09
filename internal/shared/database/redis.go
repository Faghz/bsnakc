package database

import (
	"context"
	"fmt"
	"time"

	"github.com/elzestia/go-boilerplate/internal/shared/config"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a new Redis client
// If otelCfg.Enabled is true, the client will be instrumented with OpenTelemetry
func NewRedisClient(cfg *config.RedisConfig, otelCfg *config.OTelConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Add OpenTelemetry instrumentation if enabled
	if otelCfg.Enabled {
		if err := redisotel.InstrumentTracing(client); err != nil {
			return nil, fmt.Errorf("failed to instrument Redis client: %w", err)
		}
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return client, nil
}
