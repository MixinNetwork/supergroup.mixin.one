package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/interceptors"
	"github.com/MixinNetwork/supergroup.mixin.one/models"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
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
					if re.Match(message.Data) {
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
	for {
		ids, err := models.ExpiredDistributedMessageIds(ctx)
		if err != nil {
			session.Logger(ctx).Errorf("cleanUpDistributedMessages ERROR: %+v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if len(ids) == 0 {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if err := models.CleanUpExpiredDistributedMessages(ctx, ids); err != nil {
			session.Logger(ctx).Errorf("cleanUpDistributedMessages ERROR: %+v", err)
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func loopPendingDistributeMessages(ctx context.Context, conn *websocket.Conn, mc *MessageContext) error {
	defer conn.Close()

	magnifier := 10
	for {
		messages, err := models.PendingDistributedMessages(ctx, int64(magnifier*models.PendingDistributedMessageLimit))
		if err != nil {
			session.Logger(ctx).Errorf("loopPendingMessages Error: %+v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		l := len(messages)
		m, r := l/magnifier, l%magnifier
		if l > 0 {
			if l <= magnifier {
				if err := sendDistributedMessges(ctx, mc, messages); err != nil {
					session.Logger(ctx).Errorf("sendDistributedMessges Error: %+v", err)
					return err
				}
			} else {
				var wg sync.WaitGroup
				var errs []error
				for i := 0; i < magnifier; i++ {
					wg.Add(1)
					go func(i int) {
						defer wg.Done()
						start, end := i*m, (i+1)*m
						if i < r {
							start, end = start+i, end+i+1
						} else {
							start, end = start+r, end+r
						}
						if i+1 == magnifier {
							end = l
						}
						if err := sendDistributedMessges(ctx, mc, messages[start:end]); err != nil {
							errs = append(errs, err)
						}
					}(i)
				}
				wg.Wait()
				if len(errs) > 0 {
					session.Logger(ctx).Errorf("WaitGroup SendDistributedMessges ERRORS size %d, first ERROR %+v", len(errs), errs[0])
					return session.BlazeServerError(ctx, errors.New(fmt.Sprintf("SendDistributedMessges ERRORS %d", len(errs))))
				}
			}
			err := models.WriteProperty(ctx, models.MessageQueueCheckpoint, messages[len(messages)-1].UpdatedAt.Format(time.RFC3339Nano))
			if err != nil {
				session.Logger(ctx).Errorf("WriteProperty ERROR: %+v", err)
				return err
			}
		}
		if len(messages) < models.PendingDistributedMessageLimit {
			time.Sleep(500 * time.Millisecond)
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
			"data":              base64.StdEncoding.EncodeToString(message.Data),
			"representative_id": message.UserId,
			"created_at":        message.CreatedAt,
			"updated_at":        message.UpdatedAt,
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
	err := json.Unmarshal(message.Data, &a)
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
	req, err := http.NewRequest(http.MethodGet, attachment.ViewURL, nil)
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
