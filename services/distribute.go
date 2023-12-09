package services

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client/v2"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/models"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

func distribute(ctx context.Context) {
	limit := int64(80)
	system := config.AppConfig.System
	shards := make([]string, system.MessageShardSize)
	for i := int64(0); i < system.MessageShardSize; i++ {
		shard := shardId(system.MessageShardModifier, i)
		shards[i] = shard
		go pendingActiveDistributedMessages(ctx, shard, limit)
	}
	for {
		count, err := models.ClearUpExpiredDistributedMessages(ctx, shards)
		if err != nil {
			session.Logger(ctx).Errorf("ClearUpExpiredDistributedMessages ERROR: %+v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if count < 100 {
			time.Sleep(time.Minute)
		}
	}
}

func pendingActiveDistributedMessages(ctx context.Context, shard string, limit int64) {
	for {
		messages, err := models.PendingActiveDistributedMessages(ctx, shard, limit)
		if err != nil {
			session.Logger(ctx).Errorf("PendingActiveDistributedMessages ERROR: %+v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if len(messages) < 1 {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		results, err := sendDistributedMessges(ctx, shard, messages)
		if err != nil {
			session.Logger(ctx).Errorf("PendingActiveDistributedMessages sendDistributedMessges ERROR: %+v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		var delivered []models.DistributedMessageResult
		var sessions []*models.Session
		for _, m := range results {
			if m.State == "SUCCESS" {
				var sessions []string
				for _, s := range m.Sessions {
					sessions = append(sessions, s.SessionID)
				}
				delivered = append(delivered, models.DistributedMessageResult{
					MessageID: m.MessageID,
					State:     models.MessageStatusDelivered,
					Sessions:  strings.Join(sessions, ","),
				})
			}
			if m.State == "FAILED" {
				if len(m.Sessions) == 0 {
					query := "UPDATE users SET subscribed_at=$1 WHERE user_id=$2"
					if _, err := session.Database(ctx).ExecContext(ctx, query, time.Time{}, m.RecipientID); err != nil {
						log.Println("UPDATE users err", err)
						continue
					}
					query = "delete from distributed_messages where recipient_id = $1 and status = 'SENT'"
					if _, err := session.Database(ctx).ExecContext(ctx, query, m.RecipientID); err != nil {
						log.Println("delete distributed_messages err", err)
						continue
					}
					continue
				}
				for _, s := range m.Sessions {
					sessions = append(sessions, &models.Session{
						UserID:    m.RecipientID,
						SessionID: s.SessionID,
						PublicKey: s.PublicKey,
						UpdatedAt: time.Now(),
					})
				}
				sessions = append(sessions, &models.Session{
					UserID: m.RecipientID,
				})
			}
		}
		err = models.UpdateDeliveredMessagesStatus(ctx, delivered)
		if err != nil {
			session.Logger(ctx).Errorf("PendingActiveDistributedMessages UpdateMessagesStatus ERROR: %+v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		err = models.SyncSession(ctx, sessions)
		if err != nil {
			session.Logger(ctx).Errorf("PendingActiveDistributedMessages SyncSession ERROR: %+v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
	}
}

type Message struct {
	MessageID   string `json:"message_id"`
	RecipientID string `json:"recipient_id"`
	State       string `json:"state"`
	Sessions    []struct {
		SessionID string `json:"session_id"`
		PublicKey string `json:"public_key"`
	} `json:"sessions"`
}

func sendDistributedMessges(ctx context.Context, key string, messages []*models.DistributedMessage) ([]*Message, error) {
	var userIDs []string
	for _, m := range messages {
		userIDs = append(userIDs, m.RecipientId)
	}
	sessionSet, err := models.ReadSessionSetByUsers(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	var body []map[string]interface{}
	for _, message := range messages {
		if message.UserId == config.AppConfig.Mixin.ClientId {
			message.UserId = ""
		}
		if message.Category == models.MessageCategoryMessageRecall {
			message.UserId = ""
		}
		m := map[string]interface{}{
			"conversation_id":   message.ConversationId,
			"recipient_id":      message.RecipientId,
			"message_id":        message.MessageId,
			"quote_message_id":  message.QuoteMessageId,
			"category":          message.Category,
			"data_base64":       message.Data,
			"silent":            message.Silent,
			"representative_id": message.UserId,
			"created_at":        message.CreatedAt,
			"updated_at":        message.CreatedAt,
		}
		recipient := sessionSet[message.RecipientId]
		category := message.ReadCategory(recipient)
		m["category"] = category
		if recipient != nil {
			m["checksum"] = models.GenerateUserChecksum(recipient.Sessions)
			var sessions []map[string]string
			for _, s := range recipient.Sessions {
				sessions = append(sessions, map[string]string{"session_id": s.SessionID})
			}
			m["recipient_sessions"] = sessions
			if strings.Contains(category, "ENCRYPTED") {
				data, err := models.EncryptMessageData(message.Data, recipient.Sessions)
				if err != nil {
					return nil, err
				}
				m["data_base64"] = data
			}
		}

		body = append(body, m)
	}

	msgs, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	mixin := config.AppConfig.Mixin
	accessToken, err := bot.SignAuthenticationToken(mixin.ClientId, mixin.SessionId, mixin.SessionKey, "POST", "/encrypted_messages", string(msgs))
	if err != nil {
		return nil, err
	}
	data, err := request(ctx, key, "POST", "/encrypted_messages", msgs, accessToken)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data  []*Message `json:"data"`
		Error bot.Error  `json:"error"`
	}
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Error.Code > 0 {
		return nil, resp.Error
	}
	return resp.Data, nil
}

var httpPool map[string]*http.Client = make(map[string]*http.Client, 0)

func request(ctx context.Context, key, method, path string, body []byte, accessToken string) ([]byte, error) {
	if httpPool[key] == nil {
		httpPool[key] = &http.Client{Timeout: 6 * time.Second}
	}
	cfg := config.AppConfig
	url := cfg.Service.APIRoot[cfg.Service.Retry%len(cfg.Service.APIRoot)]
	req, err := http.NewRequest(method, url+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := httpPool[key].Do(req)
	if err != nil {
		cfg.Service.Retry++
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		cfg.Service.Retry++
		return nil, bot.ServerError(ctx, nil)
	}
	return ioutil.ReadAll(resp.Body)
}
