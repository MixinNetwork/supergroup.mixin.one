package models

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

const (
	ProhibitedMessage = "prohibited-message-property"
)

const properties_DDL = `
CREATE TABLE IF NOT EXISTS properties (
	name               VARCHAR(512) PRIMARY KEY,
	value              VARCHAR(1024) NOT NULL,
	created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
`

var propertiesColumns = []string{"name", "value", "created_at"}

func (p *Property) values() []interface{} {
	return []interface{}{p.Name, p.Value, p.CreatedAt}
}

func propertyFromRow(row durable.Row) (*Property, error) {
	var p Property
	err := row.Scan(&p.Name, &p.Value, &p.CreatedAt)
	return &p, err
}

type Property struct {
	Name      string
	Value     string
	CreatedAt time.Time
}

func CreateProperty(ctx context.Context, name string, value bool) (*Property, error) {
	v := config.AppConfig.System.ProhibitedMessageEnabled
	if v {
		v = value
	}
	property := &Property{
		Name:      name,
		Value:     fmt.Sprint(v),
		CreatedAt: time.Now(),
	}
	params, positions := compileTableQuery(propertiesColumns)
	query := fmt.Sprintf("INSERT INTO properties (%s) VALUES (%s) ON CONFLICT (name) DO UPDATE SET value=EXCLUDED.value", params, positions)
	session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, query, property.values()...)
		if err != nil {
			return err
		}
		data := config.AppConfig
		if data.System.ProhibitedMessageEnabled {
			text := data.MessageTemplate.MessageAllow
			if value {
				text = data.MessageTemplate.MessageProhibit
			}
			return createSystemMessage(ctx, tx, "PLAIN_TEXT", base64.StdEncoding.EncodeToString([]byte(text)))
		}
		return nil
	})
	_, err := session.Database(ctx).ExecContext(ctx, query, property.values()...)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return property, nil
}

func ReadProperty(ctx context.Context, name string) (*Property, error) {
	query := fmt.Sprintf("SELECT %s FROM properties WHERE name=$1", strings.Join(propertiesColumns, ","))
	row := session.Database(ctx).QueryRowContext(ctx, query, name)
	property, err := propertyFromRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return property, nil
}

func readPropertyAsBool(ctx context.Context, tx *sql.Tx, name string) (bool, error) {
	query := fmt.Sprintf("SELECT %s FROM properties WHERE name=$1", strings.Join(propertiesColumns, ","))
	row := tx.QueryRowContext(ctx, query, name)
	property, err := propertyFromRow(row)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return property.Value == "true", nil
}

func ReadProhibitedProperty(ctx context.Context) (bool, error) {
	if config.AppConfig.System.ProhibitedMessageEnabled {
		var b bool
		err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
			var err error
			b, err = readPropertyAsBool(ctx, tx, ProhibitedMessage)
			return err
		})
		if err != nil {
			return false, session.TransactionError(ctx, err)
		}
		return b, nil
	}
	return false, nil
}

func readProhibitedStatus(ctx context.Context, tx *sql.Tx) (bool, error) {
	if config.AppConfig.System.ProhibitedMessageEnabled {
		return readPropertyAsBool(ctx, tx, ProhibitedMessage)
	}
	return false, nil
}
