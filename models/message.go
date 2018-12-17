package models

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"google.golang.org/api/iterator"
)

const (
	MessageStatePending = "pending"
	MessageStateSuccess = "success"
)

const messages_DDL = `
CREATE TABLE messages (
	message_id	          STRING(36) NOT NULL,
	user_id	              STRING(36) NOT NULL,
	category              STRING(512) NOT NULL,
	data                  BYTES(MAX) NOT NULL,
	created_at            TIMESTAMP NOT NULL,
	updated_at            TIMESTAMP NOT NULL,
	state                 STRING(128) NOT NULL,
	last_distribute_at    TIMESTAMP NOT NULL,
) PRIMARY KEY(message_id);

CREATE INDEX messages_by_state_updated ON messages(state, updated_at);
`

var messagesCols = []string{"message_id", "user_id", "category", "data", "created_at", "updated_at", "state", "last_distribute_at"}

func (m *Message) values() []interface{} {
	return []interface{}{m.MessageId, m.UserId, m.Category, m.Data, m.CreatedAt, m.UpdatedAt, m.State, m.LastDistributeAt}
}

type Message struct {
	MessageId        string
	UserId           string
	Category         string
	Data             []byte
	CreatedAt        time.Time
	UpdatedAt        time.Time
	State            string
	LastDistributeAt time.Time
}

func CreateMessage(ctx context.Context, messageId, userId, category string, data []byte, createdAt, updatedAt time.Time) (*Message, error) {
	if len(data) > 5*1024 || category == "PLAIN_AUDIO" {
		return nil, nil
	}
	message := &Message{
		MessageId: messageId,
		UserId:    userId,
		Category:  category,
		Data:      data,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		State:     MessageStatePending,
	}
	if err := session.Database(ctx).Apply(ctx, []*spanner.Mutation{
		spanner.InsertOrUpdate("messages", messagesCols, message.values()),
	}, "messages", "INSERT", "CreateMessage"); err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return message, nil
}

func PendingMessages(ctx context.Context, limit int64) ([]*Message, error) {
	var messages []*Message
	txn := session.Database(ctx).ReadOnlyTransaction()
	defer txn.Close()

	query := fmt.Sprintf("SELECT message_id FROM messages@{FORCE_INDEX=messages_by_state_updated} WHERE state=@state ORDER BY state,updated_at LIMIT %d", limit)
	params := map[string]interface{}{"state": MessageStatePending}
	ids, err := readCollectionIds(ctx, txn, query, params)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	if len(ids) == 0 {
		return nil, nil
	}

	it := txn.Query(ctx, spanner.Statement{
		SQL:    fmt.Sprintf("SELECT %s FROM messages WHERE message_id IN UNNEST(@message_ids)", strings.Join(messagesCols, ",")),
		Params: map[string]interface{}{"message_ids": ids},
	})
	defer it.Stop()

	for {
		row, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return messages, session.TransactionError(ctx, err)
		}
		message, err := messageFromRow(row)
		if err != nil {
			return messages, session.TransactionError(ctx, err)
		}
		messages = append(messages, message)
	}
	sort.Slice(messages, func(i, j int) bool { return messages[i].UpdatedAt.Before(messages[j].UpdatedAt) })
	return messages, nil
}

func messageFromRow(row *spanner.Row) (*Message, error) {
	var m Message
	err := row.Columns(&m.MessageId, &m.UserId, &m.Category, &m.Data, &m.CreatedAt, &m.UpdatedAt, &m.State, &m.LastDistributeAt)
	return &m, err
}
