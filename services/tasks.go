package services

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/models"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/gofrs/uuid"
)

func loopPendingMessages(ctx context.Context) {
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

func loopPendingSuccessMessages(ctx context.Context) {
	for {
		count, err := models.LoopClearUpSuccessMessages(ctx)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			session.Logger(ctx).Errorf("PendingMessages ERROR: %+v", err)
			continue
		}
		if count < 100 {
			time.Sleep(10 * time.Minute)
		}
	}
}

func sendTextMessage(ctx context.Context, mc *MessageContext, conversationId, label string, timer *time.Timer, drained *bool) error {
	params := map[string]interface{}{
		"conversation_id": conversationId,
		"message_id":      bot.UuidNewV4().String(),
		"category":        "PLAIN_TEXT",
		"data":            base64.StdEncoding.EncodeToString([]byte(label)),
	}
	err := writeMessageAndWait(ctx, mc, "CREATE_MESSAGE", params, timer, drained)
	if err != nil {
		return session.BlazeServerError(ctx, err)
	}
	return nil
}

func sendAppButton(ctx context.Context, mc *MessageContext, label, conversationId, action string, timer *time.Timer, drained *bool) error {
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
	err = writeMessageAndWait(ctx, mc, "CREATE_MESSAGE", params, timer, drained)
	if err != nil {
		return session.BlazeServerError(ctx, err)
	}
	return nil
}

func loopInactiveUsers(ctx context.Context) {
	for {
		users, err := models.LoopingInactiveUsers(ctx)
		if err != nil {
			time.Sleep(time.Second)
			session.Logger(ctx).Errorf("LoopingInactiveUsers ERROR: %+v", err)
			continue
		}

		for _, user := range users {
			err = user.Hibernate(ctx)
			if err != nil {
				time.Sleep(time.Second)
				session.Logger(ctx).Errorf("Hibernate ERROR: %+v", err)
				continue
			}
		}
		time.Sleep(time.Hour)
	}
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
