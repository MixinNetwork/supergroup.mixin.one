package durable

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type Database struct {
	*sql.DB
}

func NewDatabase(ctx context.Context, db *sql.DB) (*Database, error) {
	return &Database{db}, db.PingContext(ctx)
}

func (d *Database) RunInTransaction(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context, tx *sql.Tx) error) error {
	tx, err := d.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	if err := fn(ctx, tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return tx.Commit()
}

type Row interface {
	Scan(dest ...interface{}) error
}
