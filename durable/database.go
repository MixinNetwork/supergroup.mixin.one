package durable

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"

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

func PrepareQuery(query string, fields []string) string {
	var columns, arguments bytes.Buffer
	for i, f := range fields {
		if i != 0 {
			columns.WriteString(",")
			arguments.WriteString(",")
		}
		columns.WriteString(f)
		arguments.WriteString(fmt.Sprintf("$%d", i+1))
	}
	return fmt.Sprintf(query, columns.String(), arguments.String())
}

type Row interface {
	Scan(dest ...interface{}) error
}
