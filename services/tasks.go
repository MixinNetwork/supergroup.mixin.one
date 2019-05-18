package services

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/interceptors"
	"github.com/MixinNetwork/supergroup.mixin.one/models"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"mvdan.cc/xurls"
)

type Attachment struct {
	AttachmentId string `json:"attachment_id"`
}

func loopPendingMessage(ctx context.Context) {
	limit := 5
	re := xurls.Relaxed()
	for {
		messages, err := models.PendingMessages(ctx, int64(limit))
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			session.Logger(ctx).Errorf("PendingMessages ERROR: %+v", err)
			continue
		}
		for _, message := range messages {
			if !config.Operators[message.UserId] {
				if config.DetectLinkEnabled && message.Category == "PLAIN_TEXT" {
					data, err := base64.StdEncoding.DecodeString(message.Data)
					if err != nil {
						session.Logger(ctx).Errorf("DetectLink ERROR: %+v", err)
					}
					if re.Match(data) {
						if err := message.Leapfrog(ctx, "Message contains link"); err != nil {
							time.Sleep(500 * time.Millisecond)
							session.Logger(ctx).Errorf("PendingMessages ERROR: %+v", err)
						}
						continue
					}
				}
				if config.DetectImageEnabled && message.Category == "PLAIN_IMAGE" {
					if b, reason := validateMessage(ctx, message); !b {
						if err := message.Leapfrog(ctx, reason); err != nil {
							time.Sleep(500 * time.Millisecond)
							session.Logger(ctx).Errorf("PendingMessages ERROR: %+v", err)
						}
						continue
					}
				}
			}
			if err := message.Distribute(ctx); err != nil {
				time.Sleep(500 * time.Millisecond)
				session.Logger(ctx).Errorf("PendingMessages ERROR: %+v", err)
				continue
			}
		}
		if len(messages) < limit {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func cleanUpDistributedMessages(ctx context.Context) {
	limit := int64(100)
	for {
		count, err := models.CleanUpExpiredDistributedMessages(ctx, limit)
		if err != nil {
			session.Logger(ctx).Errorf("cleanUpDistributedMessages ERROR: %+v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if count < 100 {
			time.Sleep(10 * time.Second)
			continue
		}
	}
}

func loopPendingDistributeMessages(ctx context.Context, conn *websocket.Conn, mc *MessageContext) error {
	defer conn.Close()

	limit := int64(20)
	for i := int64(0); i < config.MessageShardSize; i++ {
		shard := shardId(config.MessageShardModifier, i)
		go pendingDistributedMessages(ctx, shard, limit, mc)
	}
	err := make(chan error)
	return <-err
}

func pendingDistributedMessages(ctx context.Context, shard string, limit int64, mc *MessageContext) {
	for {
		messages, err := models.PendingDistributedMessages(ctx, shard, limit)
		if err != nil {
			session.Logger(ctx).Errorf("PendingDistributedMessages ERROR: %+v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if len(messages) < 1 {
			time.Sleep(2 * time.Second)
			continue
		}
		err = sendDistributedMessges(ctx, mc, messages)
		if err != nil {
			session.Logger(ctx).Errorf("PendingDistributedMessages sendDistributedMessges ERROR: %+v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		err = models.UpdateMessagesStatus(ctx, messages)
		if err != nil {
			session.Logger(ctx).Errorf("PendingDistributedMessages UpdateMessagesStatus ERROR: %+v", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
	}
}

func sendDistributedMessges(ctx context.Context, mc *MessageContext, messages []*models.DistributedMessage) error {
	var body []map[string]interface{}
	for _, message := range messages {
		if message.UserId == config.ClientId {
			message.UserId = ""
		}
		body = append(body, map[string]interface{}{
			"conversation_id":   message.ConversationId,
			"recipient_id":      message.RecipientId,
			"message_id":        message.MessageId,
			"category":          message.Category,
			"data":              message.Data,
			"representative_id": message.UserId,
			"created_at":        message.CreatedAt,
			"updated_at":        message.CreatedAt,
		})
	}
	err := writeMessageAndWait(ctx, mc, "CREATE_PLAIN_MESSAGES", map[string]interface{}{"messages": body})
	if err != nil {
		return session.BlazeServerError(ctx, err)
	}
	return nil
}

func sendTextMessage(ctx context.Context, mc *MessageContext, conversationId, label string) error {
	params := map[string]interface{}{
		"conversation_id": conversationId,
		"message_id":      bot.UuidNewV4().String(),
		"category":        "PLAIN_TEXT",
		"data":            base64.StdEncoding.EncodeToString([]byte(label)),
	}
	err := writeMessageAndWait(ctx, mc, "CREATE_MESSAGE", params)
	if err != nil {
		return session.BlazeServerError(ctx, err)
	}
	return nil
}

func sendAppButton(ctx context.Context, mc *MessageContext, label, conversationId, action string) error {
	btns, err := json.Marshal([]interface{}{map[string]string{
		"label":  label,
		"action": action,
		"color":  "#46B8DA",
	}})
	if err != nil {
		return session.BlazeServerError(ctx, err)
	}
	params := map[string]interface{}{
		"conversation_id": conversationId,
		"message_id":      bot.UuidNewV4().String(),
		"category":        "APP_BUTTON_GROUP",
		"data":            base64.StdEncoding.EncodeToString(btns),
	}
	err = writeMessageAndWait(ctx, mc, "CREATE_MESSAGE", params)
	if err != nil {
		return session.BlazeServerError(ctx, err)
	}
	return nil
}

func validateMessage(ctx context.Context, message *models.Message) (bool, string) {
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
	attachment, err := bot.AttachemntShow(ctx, config.ClientId, config.SessionId, config.SessionKey, a.AttachmentId)
	if err != nil {
		session.Logger(ctx).Errorf("validateMessage ERROR: %+v", err)
		return false, fmt.Sprintf("bot.AttachemntShow error: %+v, id: %s", err, a.AttachmentId)
	}

	session.Logger(ctx).Infof("validateMessage attachment ViewURL %s", attachment.ViewURL)
	req, err := http.NewRequest(http.MethodGet, strings.Replace(attachment.ViewURL, "assets.zeromesh.net", "s3.cn-north-1.amazonaws.com.cn", 0), nil)
	if err != nil {
		session.Logger(ctx).Errorf("validateMessage ERROR: %+v", err)
		return false, fmt.Sprintf("http.NewRequest error: %+v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	resp, _ := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		session.Logger(ctx).Errorf("validateMessage ERROR: %+v", err)
		return false, fmt.Sprintf("http.Do error: %+v", err)
	}
	defer resp.Body.Close()
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		session.Logger(ctx).Errorf("validateMessage StatusCode ERROR: %d", resp.StatusCode)
		return false, fmt.Sprintf("resp.StatusCode error: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		session.Logger(ctx).Errorf("validateMessage ERROR: %+v", err)
		return false, fmt.Sprintf("ioutil.ReadAll error: %+v", err)
	}
	if b, err := interceptors.CheckQRCode(ctx, data); b {
		if err != nil {
			return false, fmt.Sprintf("CheckQRCode error: %+v", err)
		}
		return false, "Image contains QR Code"
	}
	if b, err := interceptors.CheckSex(ctx, data); b {
		return false, fmt.Sprintf("CheckSex: %+v", err)
	}
	return true, ""
}

func shardId(modifier string, i int64) string {
	h := md5.New()
	h.Write([]byte(modifier))
	h.Write(new(big.Int).SetInt64(i).Bytes())
	s := h.Sum(nil)
	s[6] = (s[6] & 0x0f) | 0x30
	s[8] = (s[8] & 0x3f) | 0x80
	id, err := uuid.FromBytes(s)
	if err != nil {
		panic(err)
	}
	return id.String()
}
