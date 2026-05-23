package helper

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

func WithTx(ctx context.Context, db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	return WithTxOptions(ctx, db, nil, fn)
}

func WithTxOptions(ctx context.Context, db *sqlx.DB, opts *sql.TxOptions, fn func(*sqlx.Tx) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	tx, err := db.BeginTxx(ctx, opts)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
