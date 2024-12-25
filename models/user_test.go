package models

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client/v2"
	"github.com/stretchr/testify/assert"
)

func TestUserCRUD(t *testing.T) {
	assert := assert.New(t)
	ctx := setupTestContext()
	defer teardownTestContext(ctx)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	assert.Nil(err)
	public := base64.RawURLEncoding.EncodeToString(pub)
	private := base64.RawURLEncoding.EncodeToString(priv)
	authorizationID := bot.UuidNewV4().String()

	user, err := createUser(ctx, public, private, authorizationID, "", bot.UuidNewV4().String(), "1000", "name", "http://localhost")
	assert.Nil(err)
	assert.NotNil(user)
	assert.Equal("name", user.FullName)
	user, err = AuthenticateUserByToken(ctx, user.AuthenticationToken)
	assert.Nil(err)
	assert.NotNil(user)
	user, err = FindUser(ctx, user.UserId)
	assert.Nil(err)
	assert.NotNil(user)
	assert.True(user.SubscribedAt.Before(genesisStartedAt()))
	assert.Equal(int64(1000), user.IdentityNumber)

	err = user.UpdateProfile(ctx, "hello")
	assert.Nil(err)
	user, err = FindUser(ctx, user.UserId)
	assert.Nil(err)
	assert.Equal("hello", user.FullName)

	users, err := Subscribers(ctx, time.Time{}, 0, "")
	assert.Nil(err)
	assert.Len(users, 0)

	err = user.Payment(ctx)
	assert.Nil(err)
	err = user.Subscribe(ctx)
	assert.Nil(err)
	user, err = FindUser(ctx, user.UserId)
	assert.Nil(err)
	assert.True(user.SubscribedAt.After(time.Now().Add(-1 * time.Hour)))
	users, err = Subscribers(ctx, time.Time{}, 0, "")
	assert.Nil(err)
	assert.Len(users, 1)
	err = user.Unsubscribe(ctx)
	assert.Nil(err)
	user, err = FindUser(ctx, user.UserId)
	assert.Nil(err)
	assert.True(user.SubscribedAt.IsZero())
	users, err = Subscribers(ctx, time.Time{}, 0, "")
	assert.Nil(err)
	assert.Len(users, 0)
	count, err := SubscribersCount(ctx)
	assert.Nil(err)
	assert.Equal(int64(0), count)

	uid := bot.UuidNewV4().String()
	data := base64.RawURLEncoding.EncodeToString([]byte("hello"))
	message, err := CreateMessage(ctx, user, uid, MessageCategoryPlainText, "", data, false, time.Now(), time.Now())
	assert.Nil(err)
	assert.NotNil(message)
	err = message.Distribute(ctx)
	assert.Nil(err)

	err = user.Payment(ctx)
	assert.Nil(err)
	user, err = FindUser(ctx, user.UserId)
	assert.Nil(err)
	assert.Equal(PaymentStatePaid, user.State)
	messages, err := PendingMessages(ctx, 100)
	assert.Nil(err)
	assert.Len(messages, 1)
	dms, err := testReadDistributedMessages(ctx)
	assert.Nil(err)
	assert.Len(dms, 0)

	err = user.Payment(ctx)
	assert.Nil(err)
	user, err = FindUser(ctx, user.UserId)
	assert.Nil(err)
	assert.Equal(PaymentStatePaid, user.State)
	err = user.Subscribe(ctx)
	assert.Nil(err)
	messages, err = PendingMessages(ctx, 100)
	assert.Nil(err)
	assert.Len(messages, 1)
	count, err = SubscribersCount(ctx)
	assert.Nil(err)
	assert.Equal(int64(1), count)

	li, err := createUser(ctx, public, private, authorizationID, "", bot.UuidNewV4().String(), "1001", "name", "http://localhost")
	assert.Nil(err)
	assert.NotNil(li)
	assert.Equal("name", li.FullName)
	li, err = createUser(ctx, public, private, authorizationID, "", li.UserId, "1001", "fullname", "http://localhost")
	assert.Nil(err)
	assert.NotNil(li)
	assert.Equal("fullname", li.FullName)
	err = li.Payment(ctx)
	assert.Nil(err)
	users, err = Subscribers(ctx, user.SubscribedAt, 0, "")
	assert.Nil(err)
	assert.Len(users, 1)
	user, err = findUserByIdentityNumber(ctx, li.IdentityNumber)
	assert.Nil(err)
	assert.NotNil(user)

	li.DeleteUser(ctx, li.UserId)
	user, err = FindUser(ctx, li.UserId)
	assert.Nil(err)
	assert.NotNil(user)
	admin := &User{UserId: "e9e5b807-fa8b-455a-8dfa-b189d28310ff"}
	admin.DeleteUser(ctx, li.UserId)
	user, err = FindUser(ctx, li.UserId)
	assert.Nil(err)
	assert.Nil(user)
}

func TestUserSubscription(t *testing.T) {
	assert := assert.New(t)
	ctx := setupTestContext()
	defer teardownTestContext(ctx)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	assert.Nil(err)
	public := base64.RawURLEncoding.EncodeToString(pub)
	private := base64.RawURLEncoding.EncodeToString(priv)
	authorizationID := bot.UuidNewV4().String()

	id := bot.UuidNewV4().String()
	user, err := createUser(ctx, public, private, authorizationID, "", id, "1000", "name", "http://localhost")
	assert.Nil(err)
	assert.NotNil(user)
	assert.Equal("name", user.FullName)
	err = user.Payment(ctx)
	assert.Nil(err)
	err = user.Subscribe(ctx)
	assert.Nil(err)
	user, err = FindUser(ctx, user.UserId)
	assert.Nil(err)
	assert.True(user.SubscribedAt.After(time.Now().Add(-1 * time.Hour)))

	user, err = createUser(ctx, public, private, authorizationID, "", id, "1000", "name", "http://localhost")
	assert.Nil(err)
	assert.NotNil(user)
	assert.Equal("name", user.FullName)
	assert.True(user.SubscribedAt.After(time.Now().Add(-1 * time.Hour)))
	user, err = FindUser(ctx, user.UserId)
	assert.Nil(err)
	assert.True(user.SubscribedAt.After(time.Now().Add(-1 * time.Hour)))
}

func TestBlacklistCRUD(t *testing.T) {
	assert := assert.New(t)
	ctx := setupTestContext()
	defer teardownTestContext(ctx)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	assert.Nil(err)
	public := base64.RawURLEncoding.EncodeToString(pub)
	private := base64.RawURLEncoding.EncodeToString(priv)
	authorizationID := bot.UuidNewV4().String()

	admin := &User{UserId: "e9e5b807-fa8b-455a-8dfa-b189d28310ff"}
	id := bot.UuidNewV4().String()
	list, err := admin.CreateBlacklist(ctx, id)
	assert.Nil(err)
	assert.NotNil(list)
	list, err = ReadBlacklist(ctx, id)
	assert.Nil(err)
	assert.Nil(list)

	li, err := createUser(ctx, public, private, authorizationID, "", bot.UuidNewV4().String(), "1001", "name", "http://localhost")
	assert.Nil(err)
	assert.NotNil(li)
	list, err = admin.CreateBlacklist(ctx, li.UserId)
	assert.Nil(err)
	assert.NotNil(list)
	list, err = ReadBlacklist(ctx, li.UserId)
	assert.Nil(err)
	assert.NotNil(list)

	user, err := FindUser(ctx, li.UserId)
	assert.Nil(err)
	assert.Nil(user)
}
