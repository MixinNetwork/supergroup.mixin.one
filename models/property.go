package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

const (
	MessageQueueCheckpoint = "service-message-checkpoint"
)

const properties_DDL = `
CREATE TABLE IF NOT EXISTS properties (
	key         VARCHAR(512) PRIMARY KEY,
	value       VARCHAR(8192) NOT NULL,
	updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
`

type Property struct {
	Key       string
	Value     string
	UpdatedAt time.Time
}

func readProperty(ctx context.Context, tx *sql.Tx, key string) (string, error) {
	var p Property
	query := "SELECT key,value,updated_at FROM properties WHERE key=$1"
	err := tx.QueryRowContext(ctx, query, key).Scan(&p.Key, &p.Value, &p.UpdatedAt)
	return p.Value, err
}

func ReadProperty(ctx context.Context, key string) (string, error) {
	var v string
	err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		v, err = readProperty(ctx, tx, key)
		return err
	})
	if err != nil {
		return "", session.TransactionError(ctx, err)
	}
	return v, nil
}

func writeProperty(ctx context.Context, tx *sql.Tx, key, value string) error {
	query := "INSERT INTO properties ('key','value','updated_at') VALUES($1,$2,$3) ON CONFLICT ('key') DO UPDATE SET ('value', 'updated_at')=(EXCLUDED.value, EXCLUDED.updated_at)"
	_, err := tx.ExecContext(ctx, query, key, value, time.Now())
	return err
}

func WriteProperty(ctx context.Context, key, value string) error {
	err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return writeProperty(ctx, tx, key, value)
	})
	if err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func ReadPropertyAsOffset(ctx context.Context, key string) (time.Time, error) {
	var offset time.Time
	timestamp, err := ReadProperty(ctx, key)
	if err != nil {
		return offset, err
	}
	if timestamp != "" {
		return time.Parse(time.RFC3339Nano, timestamp)
	}
	return offset, nil
}
