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

func (d *Database) RunInTransaction(ctx context.Context, fn func(ctx context.Context, tx *sql.Tx) error) error {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(ctx, tx); err != nil {
		return tx.Rollback()
	}
	return tx.Commit()
}

type Row interface {
	Scan(dest ...interface{}) error
}
