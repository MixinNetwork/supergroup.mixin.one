package models

import (
	"bytes"
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/interceptors"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/gofrs/uuid"
	"github.com/lib/pq"
	"mvdan.cc/xurls"
)

const (
	DistributeSubscriberLimit = 100

	MessageStatusSent      = "SENT"
	MessageStatusDelivered = "DELIVERED"
)

type DistributedMessage struct {
	MessageId      string
	ConversationId string
	RecipientId    string
	UserId         string
	ParentId       string
	QuoteMessageId string
	Shard          string
	Category       string
	Data           string
	Status         string
	CreatedAt      time.Time
}

var distributedMessagesCols = []string{"message_id", "conversation_id", "recipient_id", "user_id", "parent_id", "quote_message_id", "shard", "category", "data", "status", "created_at"}

func (dm *DistributedMessage) values() []interface{} {
	return []interface{}{dm.MessageId, dm.ConversationId, dm.RecipientId, dm.UserId, dm.ParentId, dm.QuoteMessageId, dm.Shard, dm.Category, dm.Data, dm.Status, dm.CreatedAt}
}

func distributedMessageFromRow(row durable.Row) (*DistributedMessage, error) {
	var m DistributedMessage
	err := row.Scan(&m.MessageId, &m.ConversationId, &m.RecipientId, &m.UserId, &m.ParentId, &m.QuoteMessageId, &m.Shard, &m.Category, &m.Data, &m.Status, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &m, err
}

func createDistributeMessage(ctx context.Context, messageId, parentId, quoteMessageId, userId, recipientId, category, data string) (*DistributedMessage, error) {
	dm := &DistributedMessage{
		MessageId:      messageId,
		ConversationId: UniqueConversationId(config.AppConfig.Mixin.ClientId, recipientId),
		RecipientId:    recipientId,
		UserId:         userId,
		ParentId:       parentId,
		QuoteMessageId: quoteMessageId,
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
	system := config.AppConfig.System
	if !system.Operators[message.UserId] {
		if system.DetectLinkEnabled && message.Category == MessageCategoryPlainText {
			data, err := base64.StdEncoding.DecodeString(message.Data)
			if err != nil {
				return err
			}
			if xurls.Relaxed.Match(data) {
				return message.Notify(ctx, "Message contains link")
			}
		}
		if system.DetectQRCodeEnabled && message.Category == MessageCategoryPlainImage {
			b, reason := messageQRFilter(ctx, message)
			if !b {
				return message.Notify(ctx, reason)
			}
		}
	}

	var recall RecallMessage
	if message.Category == MessageCategoryMessageRecall {
		data, err := base64.StdEncoding.DecodeString(message.Data)
		if err != nil {
			return session.BadDataError(ctx)
		}
		err = json.Unmarshal(data, &recall)
		if err != nil {
			return session.BadDataError(ctx)
		}
	}
	var quote *Message
	if message.QuoteMessageId != "" {
		var err error
		quote, err = FindMessage(ctx, message.QuoteMessageId)
		if err != nil {
			return err
		}
	}
	for {
		users, err := subscribedUsers(ctx, message.LastDistributeAt, DistributeSubscriberLimit)
		if err != nil {
			return session.TransactionError(ctx, err)
		}
		messageIds := make([]string, len(users))
		for i, user := range users {
			messageIds[i] = UniqueConversationId(user.UserId, message.MessageId)
		}
		set, err := readDistributedMessagesByIds(ctx, messageIds)
		if err != nil {
			return session.TransactionError(ctx, err)
		}

		err = session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
			stmt, err := tx.PrepareContext(ctx, pq.CopyIn("distributed_messages", distributedMessagesCols...))
			if err != nil {
				return err
			}
			defer stmt.Close()
			for _, user := range users {
				message.LastDistributeAt = user.SubscribedAt
				if user.UserId == message.UserId {
					continue
				}
				messageId := UniqueConversationId(user.UserId, message.MessageId)
				if set[messageId] {
					continue
				}
				quoteMessageId := ""
				if message.QuoteMessageId != "" && quote != nil {
					quoteMessageId = UniqueConversationId(user.UserId, quote.MessageId)
					if quote.UserId == user.UserId {
						quoteMessageId = quote.MessageId
					}
				}
				if message.Category == MessageCategoryMessageRecall {
					r := RecallMessage{
						MessageId: UniqueConversationId(user.UserId, recall.MessageId),
					}
					data, err := json.Marshal(r)
					if err != nil {
						return session.BadDataError(ctx)
					}
					message.Data = base64.StdEncoding.EncodeToString(data)
				}
				conversationId := UniqueConversationId(config.AppConfig.Mixin.ClientId, user.UserId)
				shard, err := shardId(conversationId, user.UserId)
				if err != nil {
					return err
				}
				dm := &DistributedMessage{
					MessageId:      messageId,
					ConversationId: conversationId,
					RecipientId:    user.UserId,
					UserId:         message.UserId,
					ParentId:       message.MessageId,
					QuoteMessageId: quoteMessageId,
					Shard:          shard,
					Category:       message.Category,
					Data:           message.Data,
					Status:         MessageStatusSent,
					CreatedAt:      time.Now(),
				}
				_, err = stmt.Exec(dm.values()...)
				if err != nil {
					return err
				}
			}
			_, err = stmt.Exec()
			if err != nil {
				return err
			}
			if len(users) < DistributeSubscriberLimit {
				message.LastDistributeAt = time.Now()
				message.State = MessageStateSuccess
			}
			_, err = tx.ExecContext(ctx, "UPDATE messages SET (last_distribute_at, state)=($1, $2) WHERE message_id=$3", message.LastDistributeAt, message.State, message.MessageId)
			return err
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

func (message *Message) Notify(ctx context.Context, reason string) error {
	ids := make([]string, 0)
	for key, _ := range config.AppConfig.System.Operators {
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
	var i int
	for _, id := range ids {
		if id == message.UserId {
			continue
		}
		messageId := UniqueConversationId(id, message.MessageId)
		if set[messageId] {
			continue
		}
		dm, err := createDistributeMessage(ctx, messageId, message.MessageId, "", message.UserId, id, message.Category, message.Data)
		if err != nil {
			session.TransactionError(ctx, err)
		}
		if i > 0 {
			values.WriteString(",")
		}
		i += 1
		values.WriteString(distributedMessageValuesString(dm.MessageId, dm.ConversationId, dm.RecipientId, dm.UserId, dm.ParentId, dm.QuoteMessageId, dm.Shard, dm.Category, dm.Data, dm.Status))
		values.WriteString(",")

		why := fmt.Sprintf("MessageId: %s, Reason: %s", message.MessageId, reason)
		data := base64.StdEncoding.EncodeToString([]byte(why))
		values.WriteString(distributedMessageValuesString(bot.UuidNewV4().String(), dm.ConversationId, dm.RecipientId, dm.UserId, dm.ParentId, dm.QuoteMessageId, dm.Shard, MessageCategoryPlainText, data, dm.Status))
	}

	message.LastDistributeAt = time.Now()
	message.State = MessageStateSuccess
	err = session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
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

func notifyToLarge(ctx context.Context, messageId, userId, name string) error {
	err := session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, pq.CopyIn("distributed_messages", distributedMessagesCols...))
		if err != nil {
			return err
		}
		defer stmt.Close()

		why := fmt.Sprintf("MessageId: %s, Reason: data too large, From: %s", messageId, name)
		data := base64.StdEncoding.EncodeToString([]byte(why))
		for key, _ := range config.AppConfig.System.Operators {
			dm := &DistributedMessage{
				MessageId:      bot.UuidNewV4().String(),
				ConversationId: UniqueConversationId(config.AppConfig.Mixin.ClientId, key),
				RecipientId:    key,
				UserId:         config.AppConfig.Mixin.ClientId,
				ParentId:       messageId,
				QuoteMessageId: "",
				Category:       MessageCategoryPlainText,
				Data:           data,
				Status:         MessageStatusSent,
				CreatedAt:      time.Now(),
			}
			shard, err := shardId(dm.ConversationId, dm.RecipientId)
			if err != nil {
				return err
			}
			dm.Shard = shard
			_, err = stmt.Exec(dm.values()...)
			if err != nil {
				return err
			}
		}
		_, err = stmt.Exec()
		return err
	})
	if err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func createSystemDistributedMessage(ctx context.Context, tx *sql.Tx, user *User, category, data string) error {
	if len(data) == 0 {
		return nil
	}
	dm, err := createDistributeMessage(ctx, bot.UuidNewV4().String(), bot.UuidNewV4().String(), "", config.AppConfig.Mixin.ClientId, user.UserId, category, data)
	if err != nil {
		return session.TransactionError(ctx, err)
	}
	var values bytes.Buffer
	values.WriteString(distributedMessageValuesString(dm.MessageId, dm.ConversationId, dm.RecipientId, dm.UserId, dm.ParentId, dm.QuoteMessageId, dm.Shard, dm.Category, dm.Data, dm.Status))
	query := fmt.Sprintf("INSERT INTO distributed_messages (%s) VALUES %s", strings.Join(distributedMessagesCols, ","), values.String())
	_, err = tx.ExecContext(ctx, query)
	return err
}

func PendingActiveDistributedMessages(ctx context.Context, shard string, limit int64) ([]*DistributedMessage, error) {
	var messages []*DistributedMessage
	query := fmt.Sprintf("SELECT %s FROM distributed_messages WHERE shard=$1 AND status=$2 ORDER BY shard,status,created_at LIMIT $3", strings.Join(distributedMessagesCols, ","))
	rows, err := session.Database(ctx).QueryContext(ctx, query, shard, MessageStatusSent, limit)
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

func ClearUpExpiredDistributedMessages(ctx context.Context, shards []string) (int64, error) {
	query := fmt.Sprintf("DELETE FROM distributed_messages WHERE message_id IN (SELECT message_id FROM distributed_messages WHERE shard = ANY($1) AND status=$2 AND created_at<$3 LIMIT 100)")
	r, err := session.Database(ctx).ExecContext(ctx, query, pq.StringArray(shards), MessageStatusDelivered, time.Now().Add(-72*time.Hour))
	if err != nil {
		return 0, session.TransactionError(ctx, err)
	}
	return r.RowsAffected()
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

func FindDistributedMessage(ctx context.Context, id string) (*DistributedMessage, error) {
	query := fmt.Sprintf("SELECT %s FROM distributed_messages WHERE message_id=$1", strings.Join(distributedMessagesCols, ","))
	row := session.Database(ctx).QueryRowContext(ctx, query, id)
	dm, err := distributedMessageFromRow(row)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return dm, nil
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

func distributedMessageValuesString(id, conversationId, recipientId, userId, parentId, quoteMessageId, shard, category, data, status string) string {
	return fmt.Sprintf("('%s','%s','%s','%s','%s', '%s','%s','%s','%s','%s', '%s')", id, conversationId, recipientId, userId, parentId, quoteMessageId, shard, category, data, status, string(pq.FormatTimestamp(time.Now())))
}

func shardId(cid, uid string) (string, error) {
	minId, maxId := cid, uid
	if strings.Compare(cid, uid) > 0 {
		maxId, minId = cid, uid
	}
	h := md5.New()
	io.WriteString(h, minId)
	io.WriteString(h, maxId)

	b := new(big.Int).SetInt64(config.AppConfig.System.MessageShardSize)
	c := new(big.Int).SetBytes(h.Sum(nil))
	m := new(big.Int).Mod(c, b)
	h = md5.New()
	h.Write([]byte(config.AppConfig.System.MessageShardModifier))
	h.Write(m.Bytes())
	s := h.Sum(nil)
	s[6] = (s[6] & 0x0f) | 0x30
	s[8] = (s[8] & 0x3f) | 0x80
	sid, err := uuid.FromBytes(s)
	return sid.String(), err
}

type Attachment struct {
	AttachmentId string `json:"attachment_id"`
}

func messageQRFilter(ctx context.Context, message *Message) (bool, string) {
	var a Attachment
	src, err := base64.StdEncoding.DecodeString(message.Data)
	if err != nil {
		return false, "message.Data format error is not Base64"
	}
	err = json.Unmarshal(src, &a)
	if err != nil {
		session.Logger(ctx).Errorf("validateMessage ERROR: %+v", err)
		return false, "message.Data Unmarshal error"
	}
	attachment, err := bot.AttachmentShow(ctx, config.AppConfig.Mixin.ClientId, config.AppConfig.Mixin.SessionId, config.AppConfig.Mixin.SessionKey, a.AttachmentId)
	if err != nil {
		return false, fmt.Sprintf("bot.AttachemntShow error: %+v, id: %s", err, a.AttachmentId)
	}

	url := strings.Replace(attachment.ViewURL, "assets.zeromesh.net", "s3.cn-north-1.amazonaws.com.cn", 0)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return true, ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	resp, _ := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return true, ""
	}
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return true, ""
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return true, ""
	}
	if b, err := interceptors.CheckQRCode(ctx, data); b {
		if err != nil {
			return true, ""
		}
		return false, "Image contains QR Code"
	}
	return true, ""
}
