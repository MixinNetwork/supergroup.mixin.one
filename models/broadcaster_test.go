package models

import (
	"testing"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/stretchr/testify/assert"
)

func TestBroadcasterCRUD(t *testing.T) {
	assert := assert.New(t)
	ctx := setupTestContext()
	defer teardownTestContext(ctx)

	admin := &User{UserId: "e9e5b807-fa8b-455a-8dfa-b189d28310ff"}

	user, err := createUser(ctx, "accessToken", bot.UuidNewV4().String(), "1000", "name", "http://localhost")
	assert.Nil(err)
	assert.NotNil(user)
	broadcaster, err := admin.CreateBroadcaster(ctx, user.IdentityNumber)
	assert.Nil(err)
	assert.Equal(user.UserId, broadcaster.UserId)
	users, err := ReadBroadcasters(ctx)
	assert.Nil(err)
	assert.Len(users, 1)
}
