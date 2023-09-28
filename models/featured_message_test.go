package models

import (
	"encoding/base64"
	"testing"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/stretchr/testify/assert"
)

func TestFeaturedMessageCRUD(t *testing.T) {
	assert := assert.New(t)
	ctx := setupTestContext()
	defer teardownTestContext(ctx)

	id, uid := bot.UuidNewV4().String(), bot.UuidNewV4().String()
	user := &User{UserId: id, ActiveAt: time.Now()}
	data := base64.RawURLEncoding.EncodeToString([]byte("hello"))
	message, err := CreateMessage(ctx, user, uid, MessageCategoryPlainText, "", data, false, time.Now(), time.Now())
	assert.Nil(err)
	assert.NotNil(message)
	message, err = FindMessage(ctx, message.MessageId)
	assert.Nil(err)
	assert.NotNil(message)

	var adminID string
	for k, _ := range config.AppConfig.System.Operators {
		adminID = k
	}
	admin := &User{UserId: adminID}
	fm, err := admin.CreateFeaturedMessage(ctx, message.MessageId)
	assert.Nil(err)
	assert.NotNil(fm)
	fm, err = FindFeaturedMessage(ctx, fm.MessageId)
	assert.Nil(err)
	assert.NotNil(fm)

	fms, err := FindFeaturedMessages(ctx)
	assert.Nil(err)
	assert.Len(fms, 1)
}
