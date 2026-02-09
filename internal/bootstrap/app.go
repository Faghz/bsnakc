package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elzestia/go-boilerplate/internal/shared/adapters/telemetry"
	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/config"
	"github.com/elzestia/go-boilerplate/internal/shared/crypto"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// App holds all application dependencies
type App struct {
	config       *config.Config
	logger       ports.Logger
	db           *sqlx.DB
	redis        *redis.Client
	server       *echo.Echo
	otelProvider *telemetry.Provider
}

// New creates and initializes the application
func New() (*App, error) {
	// Load configuration first (needed for everything)
	cfg, err := initConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize OpenTelemetry BEFORE logger (so logger can extract trace IDs)
	// This is critical for proper context propagation
	otelProvider, err := initOTel(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenTelemetry: %w", err)
	}

	// Initialize logger (will use OTel trace context and log export if available)
	logger, err := initLogger(cfg, otelProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	logger.Info(context.Background(), "Initializing application",
		ports.String("service", cfg.OTel.ServiceName),
		ports.String("version", cfg.OTel.ServiceVersion),
		ports.String("environment", cfg.OTel.Environment),
		ports.Bool("otel_enabled", cfg.OTel.Enabled),
	)

	// Initialize database
	db, err := initDatabase(cfg, logger)
	if err != nil {
		logger.Error(context.Background(), "Failed to initialize database", ports.Error(err))
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize Redis
	redis, err := initRedis(cfg, logger)
	if err != nil {
		logger.Error(context.Background(), "Failed to initialize Redis", ports.Error(err))
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Initialize crypto services (domain-specific keys for better security isolation)
	logger.Info(context.Background(), "Initializing crypto services")

	// User domain crypto service
	userKeyManager, err := crypto.NewEnvKeyManager(cfg, "user")
	if err != nil {
		logger.Error(context.Background(), "Failed to initialize user key manager", ports.Error(err))
		return nil, fmt.Errorf("failed to initialize user key manager: %w", err)
	}
	userCryptoService := crypto.NewCryptoService(userKeyManager)

	// Identity domain crypto service
	identityKeyManager, err := crypto.NewEnvKeyManager(cfg, "identity")
	if err != nil {
		logger.Error(context.Background(), "Failed to initialize identity key manager", ports.Error(err))
		return nil, fmt.Errorf("failed to initialize identity key manager: %w", err)
	}
	identityCryptoService := crypto.NewCryptoService(identityKeyManager)

	// Initialize repositories (with domain-specific crypto services for PII encryption)
	repos := initRepositories(db, logger, userCryptoService, identityCryptoService)

	// Initialize services
	services := initServices(db, redis, repos, cfg, logger)

	// Initialize HTTP server
	server := initServer(cfg, services, logger)

	logger.Info(context.Background(), "Application initialized successfully")

	return &App{
		config:       cfg,
		logger:       logger,
		db:           db,
		redis:        redis,
		server:       server,
		otelProvider: otelProvider,
	}, nil
}

// Run starts the application
func (a *App) Run() error {
	// Start server in a goroutine
	go func() {
		addr := a.config.Server.Address()
		a.logger.Info(context.Background(), "Starting server", ports.String("address", addr))
		if err := a.server.Start(addr); err != nil {
			a.logger.Error(context.Background(), "Server error", ports.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	a.logger.Info(context.Background(), "Shutting down gracefully")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	a.logger.Info(context.Background(), "Application stopped")
	return nil
}

func (a *App) shutdown(ctx context.Context) error {
	// Shutdown in reverse order of initialization

	// 1. Shutdown HTTP server (stop accepting new requests)
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	// 2. Close Redis connection
	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			return fmt.Errorf("failed to close Redis: %w", err)
		}
	}

	// 3. Close database connection
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}

	// 4. Shutdown OpenTelemetry provider (flush remaining spans/metrics)
	if a.otelProvider != nil {
		a.logger.Info(ctx, "Shutting down OpenTelemetry provider")
		if err := a.otelProvider.Shutdown(ctx); err != nil {
			a.logger.Error(ctx, "Failed to shutdown OTel provider", ports.Error(err))
			// Don't return error - best effort shutdown
		}
	}

	return nil
}
