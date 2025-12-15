package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Executor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type DB struct {
	pool   *pgxpool.Pool
	logger zerolog.Logger
}

func New(pool *pgxpool.Pool, logger zerolog.Logger) *DB {
	return &DB{pool: pool, logger: logger}
}

func (db *DB) Exec() Executor {
	return db.pool
}

func (db *DB) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := db.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		db.logger.Error().Err(err).Msg("begin tx failed")
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		db.logger.Error().Err(err).Msg("tx rolled back")
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		db.logger.Error().Err(err).Msg("tx commit failed")
		return err
	}

	return nil
}
