package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/elzestia/go-boilerplate/internal/shared/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// txKey is the key type for transaction in context
type txKey struct{}

// NewPostgresDB creates a new PostgreSQL database connection
// If otelCfg.Enabled is true, the connection will be instrumented with OpenTelemetry
func NewPostgresDB(cfg *config.DatabaseConfig, otelCfg *config.OTelConfig) (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error

	if otelCfg.Enabled {
		// Register instrumented driver for OpenTelemetry
		driverName, err := otelsql.Register("postgres",
			otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to register instrumented driver: %w", err)
		}

		db, err = sqlx.Connect(driverName, cfg.DSN())
	} else {
		// Use standard postgres driver
		db, err = sqlx.Connect("postgres", cfg.DSN())
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// Querier interface for both *sqlx.DB and *sqlx.Tx
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// GetQuerier extracts the transaction from context if available, otherwise returns db
func GetQuerier(ctx context.Context, db *sqlx.DB) Querier {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return db
}

// WithTransaction executes a function within a database transaction
func WithTransaction(ctx context.Context, db *sqlx.DB, fn func(ctx context.Context) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Store transaction in context
	txCtx := context.WithValue(ctx, txKey{}, tx)

	// Execute function
	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction failed: %w; rollback failed: %v", err, rbErr)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
