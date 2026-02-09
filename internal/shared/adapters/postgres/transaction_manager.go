package postgres

import (
	"context"

	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/database"
	"github.com/jmoiron/sqlx"
)

type TransactionManager struct {
	db     *sqlx.DB
	logger ports.Logger
}

func NewTransactionManager(db *sqlx.DB, logger ports.Logger) ports.TransactionManager {
	return &TransactionManager{
		db:     db,
		logger: logger.With(ports.String("component", "transaction_manager")),
	}
}

func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tm.logger.Debug(ctx, "Starting transaction")
	err := database.WithTransaction(ctx, tm.db, fn)
	if err != nil {
		tm.logger.Warn(ctx, "Transaction rolled back", ports.Error(err))
		return err
	}
	tm.logger.Debug(ctx, "Transaction committed successfully")
	return nil
}
