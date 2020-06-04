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

type Property struct {
	Name      string
	Value     string
	CreatedAt time.Time
}

var propertiesColumns = []string{"name", "value", "created_at"}

func (p *Property) values() []interface{} {
	return []interface{}{p.Name, p.Value, p.CreatedAt}
}

func propertyFromRow(row durable.Row) (*Property, error) {
	var p Property
	err := row.Scan(&p.Name, &p.Value, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func CreateProperty(ctx context.Context, name string, value bool) (*Property, error) {
	property := &Property{
		Name:      name,
		Value:     fmt.Sprint(value),
		CreatedAt: time.Now(),
	}
	query := durable.PrepareQuery("INSERT INTO properties (%s) VALUES (%s) ON CONFLICT (name) DO UPDATE SET value=EXCLUDED.value", propertiesColumns)
	session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, query, property.values()...)
		if err != nil {
			return err
		}
		data := config.AppConfig
		text := data.MessageTemplate.MessageAllow
		if value {
			text = data.MessageTemplate.MessageProhibit
		}
		return createSystemMessage(ctx, tx, MessageCategoryPlainText, base64.StdEncoding.EncodeToString([]byte(text)))
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
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return property, nil
}

func readPropertyAsBool(ctx context.Context, tx *sql.Tx, name string) (bool, error) {
	query := fmt.Sprintf("SELECT %s FROM properties WHERE name=$1", strings.Join(propertiesColumns, ","))
	row := tx.QueryRowContext(ctx, query, name)
	property, err := propertyFromRow(row)
	if err != nil || property == nil {
		return false, err
	}
	return property.Value == "true", nil
}

func ReadProhibitedProperty(ctx context.Context) (bool, error) {
	var b bool
	err := session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		b, err = readPropertyAsBool(ctx, tx, ProhibitedMessage)
		return err
	})
	if err != nil {
		return false, session.TransactionError(ctx, err)
	}
	return b, nil
}

func readProhibitedStatus(ctx context.Context, tx *sql.Tx) (bool, error) {
	return readPropertyAsBool(ctx, tx, ProhibitedMessage)
}
