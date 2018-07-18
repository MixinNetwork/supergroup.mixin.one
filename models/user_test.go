package models

import (
	"testing"
	"time"

	"github.com/MixinMessenger/bot-api-go-client/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserCRUD(t *testing.T) {
	assert := assert.New(t)
	ctx := setupTestContext()
	defer teardownTestContext(ctx)

	user, err := createUser(ctx, "accessToken", uuid.NewV4().String(), "1000", "name", "http://localhost")
	assert.Nil(err)
	assert.NotNil(user)
	assert.Equal("name", user.FullName)
	user, err = AuthenticateUserByToken(ctx, user.AuthenticationToken)
	assert.Nil(err)
	assert.NotNil(user)
	user, err = FindUser(ctx, user.UserId)
	assert.Nil(err)
	assert.NotNil(user)
	assert.Equal(time.Time{}, user.SubscribedAt)
	assert.Equal(int64(1000), user.IdentityNumber)

	err = user.UpdateProfile(ctx, "hello")
	assert.Nil(err)
	user, err = findUserById(ctx, user.UserId)
	assert.Nil(err)
	assert.Equal("hello", user.FullName)

	users, err := Subscribers(ctx, time.Time{})
	assert.Nil(err)
	assert.Len(users, 0)

	err = user.Subscribe(ctx)
	assert.Nil(err)
	user, err = findUserById(ctx, user.UserId)
	assert.Nil(err)
	assert.True(user.SubscribedAt.After(time.Now().Add(-1 * time.Hour)))
	users, err = Subscribers(ctx, time.Time{})
	assert.Nil(err)
	assert.Len(users, 1)
	err = user.Unsubscribe(ctx)
	assert.Nil(err)
	user, err = findUserById(ctx, user.UserId)
	assert.Nil(err)
	assert.True(user.SubscribedAt.IsZero())
	users, err = Subscribers(ctx, time.Time{})
	assert.Nil(err)
	assert.Len(users, 0)
	count, err := SubscribersCount(ctx)
	assert.Nil(err)
	assert.Equal(int64(0), count)

	err = user.Payment(ctx)
	assert.Nil(err)
	user, err = findUserById(ctx, user.UserId)
	assert.Nil(err)
	assert.Equal(PaymentStatePaid, user.State)
	messages, err := PendingMessages(ctx, 100)
	assert.Nil(err)
	assert.Len(messages, 1)

	err = user.Payment(ctx)
	assert.Nil(err)
	user, err = findUserById(ctx, user.UserId)
	assert.Nil(err)
	assert.Equal(PaymentStatePaid, user.State)
	messages, err = PendingMessages(ctx, 100)
	assert.Nil(err)
	assert.Len(messages, 1)
	count, err = SubscribersCount(ctx)
	assert.Nil(err)
	assert.Equal(int64(1), count)

	li, err := createUser(ctx, "accessToken", uuid.NewV4().String(), "1001", "name", "http://localhost")
	assert.Nil(err)
	assert.NotNil(li)
	err = li.Payment(ctx)
	assert.Nil(err)
	users, err = Subscribers(ctx, user.SubscribedAt)
	assert.Nil(err)
	assert.Len(users, 1)

	li.DeleteUser(ctx, li.UserId)
	user, err = findUserById(ctx, li.UserId)
	assert.Nil(err)
	assert.NotNil(user)
	admin := &User{UserId: "e9e5b807-fa8b-455a-8dfa-b189d28310ff"}
	admin.DeleteUser(ctx, li.UserId)
	user, err = findUserById(ctx, li.UserId)
	assert.Nil(err)
	assert.Nil(user)
}
