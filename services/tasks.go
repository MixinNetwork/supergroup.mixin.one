package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/models"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/gorilla/websocket"
)

func loopPendingMessage(ctx context.Context) {
	limit := 5
	for {
		messages, err := models.PendingMessages(ctx, int64(limit))
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			session.Logger(ctx).Errorf("PendingMessages ERROR: %+v", err)
			continue
		}
		for _, message := range messages {
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
