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
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/MixinNetwork/supergroup.mixin.one/utils"
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
	Silent         bool
	Status         string
	CreatedAt      time.Time
}

var distributedMessagesCols = []string{"message_id", "conversation_id", "recipient_id", "user_id", "parent_id", "quote_message_id", "shard", "category", "data", "silent", "status", "created_at"}

func (dm *DistributedMessage) values() []interface{} {
	return []interface{}{dm.MessageId, dm.ConversationId, dm.RecipientId, dm.UserId, dm.ParentId, dm.QuoteMessageId, dm.Shard, dm.Category, dm.Data, dm.Silent, dm.Status, dm.CreatedAt}
}

func distributedMessageFromRow(row durable.Row) (*DistributedMessage, error) {
	var m DistributedMessage
	err := row.Scan(&m.MessageId, &m.ConversationId, &m.RecipientId, &m.UserId, &m.ParentId, &m.QuoteMessageId, &m.Shard, &m.Category, &m.Data, &m.Silent, &m.Status, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &m, err
}

func (message *Message) Distribute(ctx context.Context) error {
	prohibited, err := ReadProhibitedProperty(ctx)
	if err != nil {
		return err
	}
	system := config.AppConfig.System
	if !system.Operators[message.UserId] {
		switch message.Category {
		case MessageCategoryPlainText, MessageCategoryEncryptedText:
			if system.DetectLinkEnabled {
				data, err := base64.RawURLEncoding.DecodeString(message.Data)
				if err != nil {
					return err
				}
				if xurls.Relaxed.Match(data) {
					return message.Notify(ctx, "Message contains link")
				}
			}
		case MessageCategoryPlainImage, MessageCategoryEncryptedImage:
			if system.DetectQRCodeEnabled {
				b, reason := messageQRFilter(ctx, message)
				if !b {
					return message.Notify(ctx, reason)
				}
			}
		}
	}

	var recall RecallMessage
	if message.Category == MessageCategoryMessageRecall {
		data, err := base64.RawURLEncoding.DecodeString(message.Data)
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

	var transcripts []Transcript
	switch message.Category {
	case MessageCategoryPlainTranscript,
		MessageCategoryEncryptedTranscript:
		data, err := base64.RawURLEncoding.DecodeString(message.Data)
		if err != nil {
			return session.BadDataError(ctx)
		}
		err = json.Unmarshal(data, &transcripts)
		if err != nil {
			return session.BadDataError(ctx)
		}
	}

	for {
		users, err := subscribedUsers(ctx, message.LastDistributeAt, DistributeSubscriberLimit, message.UserId)
		if err != nil {
			return session.TransactionError(ctx, err)
		}
		messageIds := make([]string, len(users))
		for i, user := range users {
			messageIds[i] = UniqueConversationId(user.UserId, message.MessageId)
		}
		set, err := readDistributedMessageSetByIds(ctx, messageIds)
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
				if !config.AppConfig.System.Operators[message.UserId] && message.UserId != config.AppConfig.Mixin.ClientId && prohibited {
					if !config.AppConfig.System.Operators[user.UserId] {
						continue
					}
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
				switch message.Category {
				case MessageCategoryMessageRecall:
					r := RecallMessage{
						MessageId: UniqueConversationId(user.UserId, recall.MessageId),
					}
					data, err := json.Marshal(r)
					if err != nil {
						return session.BadDataError(ctx)
					}
					message.Data = base64.RawURLEncoding.EncodeToString(data)
				case MessageCategoryPlainTranscript,
					MessageCategoryEncryptedTranscript:
					for _, t := range transcripts {
						t["transcript_id"] = messageId
					}
					data, err := json.Marshal(transcripts)
					if err != nil {
						return session.BadDataError(ctx)
					}
					message.Data = base64.RawURLEncoding.EncodeToString(data)
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
					Silent:         message.Silent,
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
	set, err := readDistributedMessageSetByIds(ctx, messageIds)
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
		dm, err := buildDistributeMessage(ctx, messageId, message.MessageId, "", message.UserId, id, message.Category, message.Data, false)
		if err != nil {
			session.TransactionError(ctx, err)
		}
		if i > 0 {
			values.WriteString(",")
		}
		i += 1
		values.WriteString(distributedMessageValuesString(dm.MessageId, dm.ConversationId, dm.RecipientId, dm.UserId, dm.ParentId, dm.QuoteMessageId, dm.Shard, dm.Category, dm.Data, dm.Silent, dm.Status))
		values.WriteString(",")

		why := fmt.Sprintf("MessageId: %s, Reason: %s", message.MessageId, reason)
		data := base64.RawURLEncoding.EncodeToString([]byte(why))
		values.WriteString(distributedMessageValuesString(bot.UuidNewV4().String(), dm.ConversationId, dm.RecipientId, dm.UserId, dm.ParentId, dm.QuoteMessageId, dm.Shard, MessageCategoryPlainText, data, dm.Silent, dm.Status))
	}

	message.LastDistributeAt = time.Now()
	message.State = MessageStateSuccess
	err = session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		_, err = tx.ExecContext(ctx, "UPDATE messages SET (last_distribute_at, state)=($1, $2) WHERE message_id=$3", message.LastDistributeAt, message.State, message.MessageId)
		if err != nil {
			return err
		}
		valString := values.String()
		if valString != "" {
			query := fmt.Sprintf("INSERT INTO distributed_messages (%s) VALUES %s", strings.Join(distributedMessagesCols, ","), values.String())
			_, err = tx.ExecContext(ctx, query)
		}
		return err
	})
	if err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func notifyTooLarge(ctx context.Context, messageId, category, userId, name string) error {
	err := session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, pq.CopyIn("distributed_messages", distributedMessagesCols...))
		if err != nil {
			return err
		}
		defer stmt.Close()

		why := fmt.Sprintf("MessageId: %s, Category: %s, Reason: data too large, From: %s", messageId, category, name)
		data := base64.RawURLEncoding.EncodeToString([]byte(why))
		mixin := config.AppConfig.Mixin
		for key, _ := range config.AppConfig.System.Operators {
			dm := &DistributedMessage{
				MessageId:      bot.UuidNewV4().String(),
				ConversationId: UniqueConversationId(mixin.ClientId, key),
				RecipientId:    key,
				UserId:         mixin.ClientId,
				ParentId:       messageId,
				QuoteMessageId: "",
				Category:       MessageCategoryPlainText,
				Data:           data,
				Silent:         false,
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

func CreateSystemDistributedMessage(ctx context.Context, user *User, category, reason string) error {
	err := session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		return createSystemDistributedMessageInTx(ctx, tx, user, MessageCategoryPlainText, reason)
	})
	return err
}

func createSystemDistributedMessageInTx(ctx context.Context, tx *sql.Tx, user *User, category, reason string) error {
	if len(reason) == 0 {
		return nil
	}
	dm, err := buildDistributeMessage(ctx, bot.UuidNewV4().String(), bot.UuidNewV4().String(), "", config.AppConfig.Mixin.ClientId, user.UserId, category, reason, false)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, pq.CopyIn("distributed_messages", distributedMessagesCols...))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, dm.values()...)
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

type DistributedMessageResult struct {
	MessageID string
	State     string
	Sessions  string
}

func UpdateDeliveredMessagesStatus(ctx context.Context, result []DistributedMessageResult) error {
	if len(result) < 1 {
		return nil
	}
	var rows []string
	for _, r := range result {
		rows = append(rows, fmt.Sprintf("('%s', '%s', '%s')", r.MessageID, r.State, r.Sessions))
	}
	query := "UPDATE distributed_messages SET (status, sessions)=(m.state, m.sessions) FROM (values %s) as m(message_id, state, sessions) WHERE distributed_messages.message_id=m.message_id"
	_, err := session.Database(ctx).ExecContext(ctx, fmt.Sprintf(query, strings.Join(rows, ",")))
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

func readDistributedMessageSetByIds(ctx context.Context, ids []string) (map[string]bool, error) {
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

func distributedMessageValuesString(id, conversationId, recipientId, userId, parentId, quoteMessageId, shard, category, data string, silent bool, status string) string {
	return fmt.Sprintf("('%s','%s','%s','%s','%s', '%s','%s','%s','%s',%t,'%s','%s')", id, conversationId, recipientId, userId, parentId, quoteMessageId, shard, category, data, silent, status, string(pq.FormatTimestamp(time.Now())))
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
	src, err := base64.RawURLEncoding.DecodeString(message.Data)
	if err != nil {
		return false, "message.Data format error is not Base64"
	}
	err = json.Unmarshal(src, &a)
	if err != nil {
		session.Logger(ctx).Errorf("validateMessage ERROR: %+v", err)
		return false, "message.Data Unmarshal error"
	}
	mixin := config.AppConfig.Mixin
	attachment, err := bot.AttachmentShow(ctx, mixin.ClientId, mixin.SessionId, mixin.SessionKey, a.AttachmentId)
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
	if b, err := utils.CheckQRCode(ctx, data); b {
		if err != nil {
			return true, ""
		}
		return false, "Image contains QR Code"
	}
	return true, ""
}

func buildDistributeMessage(ctx context.Context, messageId, parentId, quoteMessageId, userId, recipientId, category, data string, silent bool) (*DistributedMessage, error) {
	dm := &DistributedMessage{
		MessageId:      messageId,
		ConversationId: UniqueConversationId(config.AppConfig.Mixin.ClientId, recipientId),
		RecipientId:    recipientId,
		UserId:         userId,
		ParentId:       parentId,
		QuoteMessageId: quoteMessageId,
		Category:       category,
		Data:           data,
		Silent:         silent,
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

func UniqueConversationId(userId, recipientId string) string {
	minId, maxId := userId, recipientId
	if strings.Compare(userId, recipientId) > 0 {
		maxId, minId = userId, recipientId
	}
	h := md5.New()
	io.WriteString(h, minId)
	io.WriteString(h, maxId)
	sum := h.Sum(nil)
	sum[6] = (sum[6] & 0x0f) | 0x30
	sum[8] = (sum[8] & 0x3f) | 0x80
	return uuid.FromBytesOrNil(sum).String()
}

func (m *DistributedMessage) ReadCategory(user *SimpleUser) string {
	if user == nil {
		return strings.Replace(m.Category, "ENCRYPTED_", "PLAIN_", -1)
	}
	switch user.Category {
	case UserCategoryPlain:
		return strings.Replace(m.Category, "ENCRYPTED_", "PLAIN_", -1)
	case UserCategoryEncrypted:
		return strings.Replace(m.Category, "PLAIN_", "ENCRYPTED_", -1)
	default:
		return m.Category
	}
}

/*
"category": "",
"caption": "",
"content": "",
"created_at": "",
"media_name": "",
"media_key": "",
"media_waveform": "",
"media_size": 0,
"media_width": 0,
"media_height": 0,
"media_duration": 0,
"media_url": "",
"media_status": "",
"media_digest": "",
"media_mime_type": "",
"media_created_at": ""
"mentions": "",
"message_id": "",
"shared_user_id": "",
"sticker_id": "",
"transcript_id": "",
"thumb_url": "",
"thumb_image": "",
"user_id": "",
"user_full_name": "",
"quote_id": "",
"quote_content": "",
*/
type Transcript map[string]interface{}
