package models

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"google.golang.org/api/iterator"
)

const (
	DistributeSubscriberLimit      = 100
	ExpiredDistributedMessageLimit = 100
	PendingDistributedMessageLimit = 20
)

const distributed_messages_DDL = `
CREATE TABLE distributed_messages (
	message_id            STRING(36) NOT NULL,
	conversation_id       STRING(36) NOT NULL,
	recipient_id          STRING(36) NOT NULL,
	user_id	              STRING(36) NOT NULL,
	category              STRING(512) NOT NULL,
	data                  BYTES(MAX) NOT NULL,
	created_at            TIMESTAMP NOT NULL,
	updated_at            TIMESTAMP NOT NULL,
) PRIMARY KEY(message_id);

CREATE INDEX distributed_messages_by_updated ON distributed_messages(updated_at);
`

var distributedMessagesCols = []string{"message_id", "conversation_id", "recipient_id", "user_id", "category", "data", "created_at", "updated_at"}

func (dm *DistributedMessage) values() []interface{} {
	return []interface{}{dm.MessageId, dm.ConversationId, dm.RecipientId, dm.UserId, dm.Category, dm.Data, dm.CreatedAt, dm.UpdatedAt}
}

type DistributedMessage struct {
	MessageId      string
	ConversationId string
	RecipientId    string
	UserId         string
	Category       string
	Data           []byte
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func createDistributeMessage(ctx context.Context, messageId, userId, recipientId, category string, data []byte) *spanner.Mutation {
	t := time.Now()
	dm := &DistributedMessage{
		MessageId:      messageId,
		ConversationId: UniqueConversationId(config.ClientId, recipientId),
		RecipientId:    recipientId,
		UserId:         userId,
		Category:       category,
		Data:           data,
		CreatedAt:      t,
		UpdatedAt:      t,
	}
	return spanner.Insert("distributed_messages", distributedMessagesCols, dm.values())
}

func (message *Message) Distribute(ctx context.Context) error {
	for {
		ids, subscribedAt, err := subscribedUserIds(ctx, message.LastDistributeAt, DistributeSubscriberLimit)
		if err != nil {
			return session.TransactionError(ctx, err)
		}
		mutations := []*spanner.Mutation{}
		messageIds := make([]string, len(ids))
		for i, id := range ids {
			messageIds[i] = UniqueConversationId(id, message.MessageId)
		}
		set, err := readDistributedMessagesByIds(ctx, messageIds)
		if err != nil {
			return session.TransactionError(ctx, err)
		}
		for _, id := range ids {
			if id == message.UserId {
				continue
			}
			messageId := UniqueConversationId(id, message.MessageId)
			if set[messageId] {
				continue
			}
			mutations = append(mutations, createDistributeMessage(ctx, messageId, message.UserId, id, message.Category, message.Data))
		}
		if len(ids) < DistributeSubscriberLimit {
			message.LastDistributeAt = time.Now()
			message.State = MessageStateSuccess
			mutations = append(mutations, spanner.Update("messages", []string{"message_id", "state", "last_distribute_at"}, []interface{}{message.MessageId, message.State, message.LastDistributeAt}))
		} else {
			message.LastDistributeAt = subscribedAt
			mutations = append(mutations, spanner.Update("messages", []string{"message_id", "last_distribute_at"}, []interface{}{message.MessageId, message.LastDistributeAt}))
		}
		if err := session.Database(ctx).Apply(ctx, mutations, "distributed_messages", "INSERT", "DistributeMessage"); err != nil {
			return session.TransactionError(ctx, err)
		}
		if message.State == MessageStateSuccess {
			break
		}
	}
	return nil
}

func PendingDistributedMessages(ctx context.Context, limit int64) ([]*DistributedMessage, error) {
	var messages []*DistributedMessage
	txn := session.Database(ctx).ReadOnlyTransaction()
	defer txn.Close()

	offset, err := ReadPropertyAsOffset(ctx, MessageQueueCheckpoint)
	if err != nil {
		return messages, err
	}
	query := fmt.Sprintf("SELECT message_id FROM distributed_messages@{FORCE_INDEX=distributed_messages_by_updated} WHERE updated_at>@updated_at ORDER BY updated_at LIMIT %d", limit)
	params := map[string]interface{}{"updated_at": offset}
	ids, err := readCollectionIds(ctx, txn, query, params)
	if err != nil || len(ids) == 0 {
		return messages, err
	}

	it := txn.Query(ctx, spanner.Statement{
		SQL:    fmt.Sprintf("SELECT %s FROM distributed_messages WHERE message_id IN UNNEST(@message_ids)", strings.Join(distributedMessagesCols, ",")),
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

		message, err := distributedMessageFromRow(row)
		if err != nil {
			return messages, session.TransactionError(ctx, err)
		}
		messages = append(messages, message)
	}
	sort.Slice(messages, func(i, j int) bool { return messages[i].UpdatedAt.Before(messages[j].UpdatedAt) })
	return messages, nil
}

func ExpiredDistributedMessageIds(ctx context.Context) ([]string, error) {
	txn := session.Database(ctx).ReadOnlyTransaction()
	defer txn.Close()

	var offset time.Time
	timestamp, err := readProperty(ctx, txn, MessageQueueCheckpoint)
	if err != nil {
		return nil, err
	}
	if timestamp != "" {
		offset, err = time.Parse(time.RFC3339Nano, timestamp)
	}
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SELECT message_id FROM distributed_messages@{FORCE_INDEX=distributed_messages_by_updated} WHERE updated_at<@updated_at ORDER BY updated_at LIMIT %d", ExpiredDistributedMessageLimit)
	params := map[string]interface{}{"updated_at": offset}
	return readCollectionIds(ctx, txn, query, params)
}

func CleanUpExpiredDistributedMessages(ctx context.Context, ids []string) error {
	var keySets []spanner.KeySet
	for _, id := range ids {
		keySets = append(keySets, spanner.Key{id})
	}
	if err := session.Database(ctx).Apply(ctx, []*spanner.Mutation{
		spanner.Delete("distributed_messages", spanner.KeySets(keySets...)),
	}, "distributed_messages", "DELETE", "DeleteDistributedMessage"); err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func readDistributedMessagesByIds(ctx context.Context, ids []string) (map[string]bool, error) {
	stmt := spanner.Statement{
		SQL:    "SELECT message_id FROM distributed_messages WHERE message_id IN UNNEST(@ids)",
		Params: map[string]interface{}{"ids": ids},
	}
	it := session.Database(ctx).Query(ctx, stmt, "distributed_messages", "SELECT")
	defer it.Stop()

	set := make(map[string]bool)
	for {
		row, err := it.Next()
		if err == iterator.Done {
			return set, nil
		} else if err != nil {
			return nil, err
		}
		var id string
		if err := row.Columns(&id); err != nil {
			return nil, err
		}
		set[id] = true
	}
}

func distributedMessageFromRow(row *spanner.Row) (*DistributedMessage, error) {
	var m DistributedMessage
	err := row.Columns(&m.MessageId, &m.ConversationId, &m.RecipientId, &m.UserId, &m.Category, &m.Data, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}
