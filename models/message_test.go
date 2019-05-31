package models

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMessageCRUD(t *testing.T) {
	assert := assert.New(t)
	ctx := setupTestContext()
	defer teardownTestContext(ctx)

	id, uid := bot.UuidNewV4().String(), bot.UuidNewV4().String()
	user := &User{UserId: id, ActiveAt: time.Now()}
	data := base64.StdEncoding.EncodeToString([]byte("hello"))
	message, err := CreateMessage(ctx, user, uid, "PLAIN_TEXT", "", data, time.Now(), time.Now())
	assert.Nil(err)
	assert.NotNil(message)
	message, err = FindMessage(ctx, message.MessageId)
	assert.Nil(err)
	assert.NotNil(message)
	message, err = FindMessage(ctx, bot.UuidNewV4().String())
	assert.Nil(err)
	assert.Nil(message)
	message, err = CreateMessage(ctx, user, uid, "PLAIN_IMAGE", "", data, time.Now(), time.Now())
	assert.Nil(err)
	assert.NotNil(message)
	assert.Equal("PLAIN_IMAGE", message.Category)
	assert.True(message.LastDistributeAt.Equal(genesisStartedAt()))
	messages, err := PendingMessages(ctx, 100)
	assert.Nil(err)
	assert.Len(messages, 1)

	message, err = CreateMessage(ctx, &User{UserId: bot.UuidNewV4().String(), ActiveAt: time.Now()}, bot.UuidNewV4().String(), "PLAIN_TEXT", "", data, time.Now(), time.Now())
	assert.Nil(err)
	assert.NotNil(message)
	assert.Equal("PLAIN_TEXT", message.Category)
	assert.True(message.LastDistributeAt.Equal(genesisStartedAt()))
	message, err = testReadMessage(ctx, message.MessageId)
	assert.Nil(err)
	assert.NotNil(message)
	assert.Equal("PLAIN_TEXT", message.Category)
	assert.True(message.LastDistributeAt.Equal(genesisStartedAt()))

	messages, err = PendingMessages(ctx, 100)
	assert.Nil(err)
	assert.Len(messages, 2)

	user, err = createUser(ctx, "accessToken", bot.UuidNewV4().String(), "10000", "name", "http://localhost")
	assert.Nil(err)
	assert.NotNil(user)
	users, err := subscribedUsers(ctx, message.LastDistributeAt, 100)
	assert.Nil(err)
	assert.Len(users, 0)
	err = user.Subscribe(ctx)
	assert.Nil(err)
	users, err = subscribedUsers(ctx, message.LastDistributeAt, 100)
	assert.Nil(err)
	assert.Len(users, 1)

	err = message.Distribute(ctx)
	assert.Nil(err)
	dms, err := testReadDistributedMessages(ctx)
	assert.Nil(err)
	assert.Len(dms, 1)
	assert.Equal(users[0].UserId, dms[0].RecipientId)
	user, err = createUser(ctx, "accessToken", bot.UuidNewV4().String(), "10001", "name", "http://localhost")
	assert.Nil(err)
	err = user.Subscribe(ctx)
	assert.Nil(err)
	users, err = subscribedUsers(ctx, message.LastDistributeAt, 100)
	assert.Nil(err)
	assert.Len(users, 1)
	messages, err = PendingMessages(ctx, 100)
	assert.Nil(err)
	assert.Len(messages, 1)
	err = messages[0].Distribute(ctx)
	assert.Nil(err)
	dms, err = testReadDistributedMessages(ctx)
	assert.Nil(err)
	assert.Len(dms, 3)
	messages, err = PendingMessages(ctx, 100)
	assert.Nil(err)
	assert.Len(messages, 0)

	err = UpdateMessagesStatus(ctx, dms)
	assert.Nil(err)
	count, err := CleanUpExpiredDistributedMessages(ctx, 100)
	assert.Nil(err)
	assert.Equal(int64(0), count)
	query := "UPDATE distributed_messages SET created_at=$1"
	_, err = session.Database(ctx).ExecContext(ctx, query, time.Now().Add(-3*time.Hour))
	assert.Nil(err)
	count, err = CleanUpExpiredDistributedMessages(ctx, 100)
	assert.Nil(err)
	assert.Equal(int64(3), count)

	message, err = CreateMessage(ctx, user, uid, "PLAIN_TEXT", "", data, time.Now(), time.Now())
	assert.Nil(err)
	assert.NotNil(message)
	err = message.Leapfrog(ctx, "ONLY TEST")
	assert.Nil(err)
	dms, err = testReadDistributedMessages(ctx)
	assert.Nil(err)
	assert.Len(dms, 0)
}

func testReadMessage(ctx context.Context, id string) (*Message, error) {
	query := fmt.Sprintf("SELECT %s FROM messages WHERE message_id=$1", strings.Join(messagesCols, ","))
	row := session.Database(ctx).QueryRowContext(ctx, query, id)
	return messageFromRow(row)
}

func testReadDistributedMessages(ctx context.Context) ([]*DistributedMessage, error) {
	limit := int64(64)
	dms := make([]*DistributedMessage, 0)
	for i := int64(0); i < config.Get().System.MessageShardSize; i++ {
		shard := testShardId(config.Get().System.MessageShardModifier, i)
		messages, err := PendingActiveDistributedMessages(ctx, shard, limit)
		if err != nil {
			return dms, err
		}
		dms = append(dms, messages...)
	}
	return dms, nil
}

func testShardId(modifier string, i int64) string {
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
