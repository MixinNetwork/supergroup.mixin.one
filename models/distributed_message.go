package models

import (
	"bytes"
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"sort"
	"strings"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/gofrs/uuid"
)

const (
	DistributeSubscriberLimit      = 100
	ExpiredDistributedMessageLimit = 100
	PendingDistributedMessageLimit = 20

	MessageStatusSent      = "SENT"
	MessageStatusDelivered = "DELIVERED"
)

const distributed_messages_DDL = `
CREATE TABLE IF NOT EXISTS distributed_messages (
	message_id            VARCHAR(36) PRIMARY KEY CHECK (message_id ~* '^[0-9a-f-]{36,36}$'),
	conversation_id       VARCHAR(36) NOT NULL CHECK (conversation_id ~* '^[0-9a-f-]{36,36}$'),
	recipient_id          VARCHAR(36) NOT NULL CHECK (recipient_id ~* '^[0-9a-f-]{36,36}$'),
	user_id               VARCHAR(36) NOT NULL CHECK (user_id ~* '^[0-9a-f-]{36,36}$'),
	parent_id             VARCHAR(36) NOT NULL CHECK (parent_id ~* '^[0-9a-f-]{36,36}$'),
	shard                 VARCHAR(36) NOT NULL,
	category              VARCHAR(512) NOT NULL,
	data                  TEXT NOT NULL,
	status                VARCHAR(512) NOT NULL,
	created_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS message_shard_status_recipientx ON distributed_messages(shard, status, recipient_id, created_at);
CREATE INDEX IF NOT EXISTS message_createdx ON distributed_messages(created_at);
CREATE INDEX IF NOT EXISTS message_status_createdx ON distributed_messages(status, created_at);
`

var distributedMessagesCols = []string{"message_id", "conversation_id", "recipient_id", "user_id", "parent_id", "shard", "category", "data", "status", "created_at"}

func (dm *DistributedMessage) values() []interface{} {
	return []interface{}{dm.MessageId, dm.ConversationId, dm.RecipientId, dm.UserId, dm.ParentId, dm.Shard, dm.Category, dm.Data, dm.Status, dm.CreatedAt}
}

type DistributedMessage struct {
	MessageId      string
	ConversationId string
	RecipientId    string
	UserId         string
	ParentId       string
	Shard          string
	Category       string
	Data           string
	Status         string
	CreatedAt      time.Time
}

func createDistributeMessage(ctx context.Context, messageId, parentId, userId, recipientId, category, data string) (*DistributedMessage, error) {
	dm := &DistributedMessage{
		MessageId:      messageId,
		ConversationId: UniqueConversationId(config.ClientId, recipientId),
		RecipientId:    recipientId,
		UserId:         userId,
		ParentId:       parentId,
		Category:       category,
		Data:           data,
		Status:         MessageStatusSent,
		CreatedAt:      time.Now(),
	}
	shard, err := shardId(dm.ConversationId, dm.RecipientId)
	if err != nil {
		return nil, err
	}
	dm.Shard = shard
	return dm, nil
}

func (message *Message) Distribute(ctx context.Context) error {
	for {
		users, err := subscribedUsers(ctx, message.LastDistributeAt, DistributeSubscriberLimit)
		if err != nil {
			return session.TransactionError(ctx, err)
		}
		var last time.Time
		var values bytes.Buffer
		if len(users) > 0 {
			messageIds := make([]string, len(users))
			for i, user := range users {
				messageIds[i] = UniqueConversationId(user.UserId, message.MessageId)
			}
			set, err := readDistributedMessagesByIds(ctx, messageIds)
			if err != nil {
				return session.TransactionError(ctx, err)
			}
			i := 0
			for _, user := range users {
				last = user.SubscribedAt
				if user.UserId == message.UserId {
					continue
				}
				messageId := UniqueConversationId(user.UserId, message.MessageId)
				if set[messageId] {
					continue
				}
				dm, err := createDistributeMessage(ctx, messageId, message.MessageId, message.UserId, user.UserId, message.Category, message.Data)
				if err != nil {
					session.TransactionError(ctx, err)
				}
				if i > 0 {
					values.WriteString(",")
				}
				i += 1
				values.WriteString(distributedMessageValuesString(dm.MessageId, dm.ConversationId, dm.RecipientId, dm.UserId, dm.ParentId, dm.Shard, dm.Category, dm.Data, dm.Status))
			}
			message.LastDistributeAt = last
		}
		if len(users) < DistributeSubscriberLimit {
			message.LastDistributeAt = time.Now()
			message.State = MessageStateSuccess
		}
		err = session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
			_, err = tx.ExecContext(ctx, "UPDATE messages SET (last_distribute_at, state)=($1, $2) WHERE message_id=$3", message.LastDistributeAt, message.State, message.MessageId)
			if err != nil {
				return err
			}
			v := values.String()
			if v != "" {
				query := fmt.Sprintf("INSERT INTO distributed_messages (%s) VALUES %s", strings.Join(distributedMessagesCols, ","), values.String())
				_, err = tx.ExecContext(ctx, query)
				return err
			}
			return nil
		})
		if err != nil {
			return session.TransactionError(ctx, err)
		}
		if message.State == MessageStateSuccess {
			break
		}
	}
	return nil
}

func (message *Message) Leapfrog(ctx context.Context, reason string) error {
	ids := make([]string, 0)
	for key, _ := range config.Operators {
		ids = append(ids, key)
	}
	messageIds := make([]string, len(ids))
	for i, id := range ids {
		messageIds[i] = UniqueConversationId(id, message.MessageId)
	}
	set, err := readDistributedMessagesByIds(ctx, messageIds)
	if err != nil {
		return session.TransactionError(ctx, err)
	}
	var values bytes.Buffer
	i := 0
	for _, id := range ids {
		if id == message.UserId {
			continue
		}
		messageId := UniqueConversationId(id, message.MessageId)
		if set[messageId] {
			continue
		}
		dm, err := createDistributeMessage(ctx, messageId, message.MessageId, message.UserId, id, message.Category, message.Data)
		if err != nil {
			session.TransactionError(ctx, err)
		}
		if i > 0 {
			values.WriteString(",")
		}
		i += 1
		values.WriteString(distributedMessageValuesString(dm.MessageId, dm.ConversationId, dm.RecipientId, dm.UserId, dm.ParentId, dm.Shard, dm.Category, dm.Data, dm.Status))

		why := fmt.Sprintf("MessageId: %s, Reason: %s", message.MessageId, reason)
		data := base64.StdEncoding.EncodeToString([]byte(why))
		values.WriteString(",")
		values.WriteString(distributedMessageValuesString(bot.UuidNewV4().String(), dm.ConversationId, dm.RecipientId, dm.UserId, dm.ParentId, dm.Shard, "PLAIN_TEXT", data, dm.Status))
	}

	message.LastDistributeAt = time.Now()
	message.State = MessageStateSuccess
	err = session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		_, err = tx.ExecContext(ctx, "UPDATE messages SET (last_distribute_at, state)=($1, $2) WHERE message_id=$3", message.LastDistributeAt, message.State, message.MessageId)
		if err != nil {
			return err
		}
		query := fmt.Sprintf("INSERT INTO distributed_messages (%s) VALUES %s", strings.Join(distributedMessagesCols, ","), values.String())
		_, err = tx.ExecContext(ctx, query)
		return err
	})
	if err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func PendingDistributedMessages(ctx context.Context, limit int64) ([]*DistributedMessage, error) {
	var messages []*DistributedMessage
	query := fmt.Sprintf("SELECT %s FROM distributed_messages WHERE status=$1 ORDER BY created_at LIMIT $2", strings.Join(distributedMessagesCols, ","))
	rows, err := session.Database(ctx).QueryContext(ctx, query, MessageStatusSent, limit)
	if err != nil {
		return messages, session.TransactionError(ctx, err)
	}
	for rows.Next() {
		m, err := distributedMessageFromRow(rows)
		if err != nil {
			return messages, session.TransactionError(ctx, err)
		}
		messages = append(messages, m)
	}
	sort.Slice(messages, func(i, j int) bool { return messages[i].CreatedAt.Before(messages[j].CreatedAt) })
	return messages, nil
}

func PendingActiveDistributedMessages(ctx context.Context, shard string, limit int64) ([]*DistributedMessage, error) {
	var messages []*DistributedMessage
	query := fmt.Sprintf("SELECT %s FROM distributed_messages WHERE shard=$1 AND status=$2 AND recipient_id IN (SELECT user_id FROM users WHERE active_at>=$3 ORDER BY active_at DESC LIMIT 500) ORDER BY created_at LIMIT $4", strings.Join(distributedMessagesCols, ","))
	rows, err := session.Database(ctx).QueryContext(ctx, query, shard, MessageStatusSent, time.Now().Add(-60*time.Minute), limit)
	if err != nil {
		return messages, session.TransactionError(ctx, err)
	}
	for rows.Next() {
		m, err := distributedMessageFromRow(rows)
		if err != nil {
			return messages, session.TransactionError(ctx, err)
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func UpdateMessagesStatus(ctx context.Context, messages []*DistributedMessage) error {
	ids := make([]string, len(messages))
	for i, m := range messages {
		ids[i] = m.MessageId
	}
	query := fmt.Sprintf("UPDATE distributed_messages SET status=$1 WHERE message_id IN ('%s')", strings.Join(ids, "','"))
	_, err := session.Database(ctx).ExecContext(ctx, query, MessageStatusDelivered)
	if err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func CleanUpExpiredDistributedMessages(ctx context.Context, limit int64) (int64, error) {
	query := fmt.Sprintf("DELETE FROM distributed_messages WHERE message_id IN (SELECT message_id FROM distributed_messages WHERE status=$1 AND created_at<$2 LIMIT $3)")
	r, err := session.Database(ctx).ExecContext(ctx, query, MessageStatusDelivered, time.Now().Add(-1*time.Hour), limit)
	if err != nil {
		return 0, session.TransactionError(ctx, err)
	}
	count, err := r.RowsAffected()
	if err != nil {
		return 0, session.TransactionError(ctx, err)
	}
	return count, nil
}

func FindDistributedMessageRecipientId(ctx context.Context, id string) (string, error) {
	query := "SELECT recipient_id FROM distributed_messages WHERE message_id=$1"
	var recipient string
	err := session.Database(ctx).QueryRowContext(ctx, query, id).Scan(&recipient)
	if err == sql.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", session.TransactionError(ctx, err)
	}
	return recipient, nil
}

func AckDistributedMessage(ctx context.Context, userId string) error {
	query := "UPDATE users SET active_at=$1 WHERE user_id=$2"
	_, err := session.Database(ctx).ExecContext(ctx, query, time.Now(), userId)
	if err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func readDistributedMessagesByIds(ctx context.Context, ids []string) (map[string]bool, error) {
	set := make(map[string]bool)
	query := fmt.Sprintf("SELECT message_id FROM distributed_messages WHERE message_id IN ('%s')", strings.Join(ids, "','"))
	rows, err := session.Database(ctx).QueryContext(ctx, query)
	if err != nil {
		return set, err
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return set, err
		}
		set[id] = true
	}
	return set, nil
}

func distributedMessageFromRow(row durable.Row) (*DistributedMessage, error) {
	var m DistributedMessage
	err := row.Scan(&m.MessageId, &m.ConversationId, &m.RecipientId, &m.UserId, &m.ParentId, &m.Shard, &m.Category, &m.Data, &m.Status, &m.CreatedAt)
	return &m, err
}

func distributedMessageValuesString(id, conversationId, recipientId, userId, parentId, shard, category, data, status string) string {
	return fmt.Sprintf("('%s','%s','%s','%s','%s','%s','%s','%s','%s', current_timestamp)", id, conversationId, recipientId, userId, parentId, shard, category, data, status)
}

func shardId(cid, uid string) (string, error) {
	minId, maxId := cid, uid
	if strings.Compare(cid, uid) > 0 {
		maxId, minId = cid, uid
	}
	h := md5.New()
	io.WriteString(h, minId)
	io.WriteString(h, maxId)

	b := new(big.Int).SetInt64(config.MessageShardSize)
	c := new(big.Int).SetBytes(h.Sum(nil))
	m := new(big.Int).Mod(c, b)
	h = md5.New()
	h.Write([]byte(config.MessageShardModifier))
	h.Write(m.Bytes())
	s := h.Sum(nil)
	s[6] = (s[6] & 0x0f) | 0x30
	s[8] = (s[8] & 0x3f) | 0x80
	sid, err := uuid.FromBytes(s)
	return sid.String(), err
}
