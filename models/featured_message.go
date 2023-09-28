package models

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

type FeaturedMessage struct {
	MessageId      string
	UserId         string
	Category       string
	QuoteMessageId string
	Data           string
	CreatedAt      time.Time
	UpdatedAt      time.Time

	FullName string
}

var featuredMessagesCols = []string{"message_id", "user_id", "category", "quote_message_id", "data", "created_at", "updated_at"}

func (m *FeaturedMessage) values() []interface{} {
	return []interface{}{m.MessageId, m.UserId, m.Category, m.QuoteMessageId, m.Data, m.CreatedAt, m.UpdatedAt}
}

func featuredMessageFromRow(row durable.Row) (*FeaturedMessage, error) {
	var m FeaturedMessage
	err := row.Scan(&m.MessageId, &m.UserId, &m.Category, &m.QuoteMessageId, &m.Data, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &m, err
}

func (user *User) CreateFeaturedMessage(ctx context.Context, messageId string) (*FeaturedMessage, error) {
	if !user.isAdmin() {
		return nil, session.ForbiddenError(ctx)
	}

	m, err := FindMessage(ctx, messageId)
	if err != nil || m == nil {
		return nil, err
	}
	fm := &FeaturedMessage{
		MessageId:      m.MessageId,
		UserId:         m.UserId,
		Category:       m.Category,
		QuoteMessageId: m.QuoteMessageId,
		Data:           m.Data,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}

	quote, err := FindMessage(ctx, fm.QuoteMessageId)
	if err != nil {
		return nil, err
	}

	query := durable.PrepareQuery("INSERT INTO featured_messages (%s) VALUES (%s) ON CONFLICT (message_id) DO NOTHING", featuredMessagesCols)
	err = session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		_, err = tx.ExecContext(ctx, query, fm.values()...)
		if err != nil {
			return err
		}
		if quote != nil {
			fmquote := &FeaturedMessage{
				MessageId:      quote.MessageId,
				UserId:         quote.UserId,
				Category:       quote.Category,
				QuoteMessageId: quote.QuoteMessageId,
				Data:           quote.Data,
				CreatedAt:      quote.CreatedAt,
				UpdatedAt:      quote.UpdatedAt,
			}
			_, err = tx.ExecContext(ctx, query, fmquote.values()...)
		}
		return err
	})

	if err != nil {
		if sessionErr, ok := err.(session.Error); ok {
			return nil, sessionErr
		}
		return nil, session.TransactionError(ctx, err)
	}
	return fm, nil
}

func FindFeaturedMessages(ctx context.Context) ([]*FeaturedMessage, error) {
	users, err := FindFeaturedMessageUsers(ctx)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SELECT %s FROM featured_messages ORDER BY created_at DESC LIMIT 2000", strings.Join(featuredMessagesCols, ","))
	rows, err := session.Database(ctx).QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var messages []*FeaturedMessage
	for rows.Next() {
		message, err := featuredMessageFromRow(rows)
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		}
		var name = "NONE"
		if users[message.UserId] != nil {
			name = users[message.UserId].FullName
		}
		message.FullName = name
		if message.Category == MessageCategoryPlainText {
			data, _ := base64.RawURLEncoding.DecodeString(message.Data)
			message.Data = string(data)
		} else {
			message.Data = ""
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func FindFeaturedMessageUsers(ctx context.Context) (map[string]*User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE user_id IN (SELECT user_id from (SELECT DISTINCT user_id, created_at FROM featured_messages ORDER BY created_at DESC LIMIT 3000) as fms)", strings.Join(usersCols, ","))
	rows, err := session.Database(ctx).QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	users := make(map[string]*User, 0)
	for rows.Next() {
		user, err := userFromRow(rows)
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		}
		users[user.UserId] = user
	}
	return users, nil
}

func FindFeaturedMessage(ctx context.Context, id string) (*FeaturedMessage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, nil
	}
	query := fmt.Sprintf("SELECT %s FROM featured_messages WHERE message_id=$1", strings.Join(featuredMessagesCols, ","))
	row := session.Database(ctx).QueryRowContext(ctx, query, id)
	message, err := featuredMessageFromRow(row)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	if message != nil {
		user, err := FindUser(ctx, message.UserId)
		if err != nil {
			return nil, err
		}
		if user != nil {
			message.FullName = user.FullName
		}
	}
	return message, nil
}
