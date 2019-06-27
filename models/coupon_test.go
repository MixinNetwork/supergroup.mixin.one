package models

import (
	"testing"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/stretchr/testify/assert"
)

func TestCouponCRUD(t *testing.T) {
	assert := assert.New(t)
	ctx := setupTestContext()
	defer teardownTestContext(ctx)

	user2, err := createUser(ctx, "accessToken", bot.UuidNewV4().String(), "11000", "name", "http://localhost")
	assert.Nil(err)
	assert.NotNil(user2)
	coupons, err := user2.Coupons(ctx)
	assert.Nil(err)
	assert.Len(coupons, 2)
	coupons, err = user2.Coupons(ctx)
	assert.Nil(err)
	assert.Len(coupons, 2)
	coupons, err = ReadCoupons(ctx)
	assert.Nil(err)
	assert.Len(coupons, 2)

	user, err := createUser(ctx, "accessToken", bot.UuidNewV4().String(), "1000", "name", "http://localhost")
	assert.Nil(err)
	assert.NotNil(user)
	coupon := coupons[0]
	coupon, err = Occupied(ctx, coupon.Code, user)
	assert.Nil(err)
	assert.NotNil(coupons)
	assert.Equal(user.UserId, coupon.OccupiedBy.String)
	coupons, err = CreateCoupons(ctx, user, 10)
	assert.NotNil(err)
	assert.Nil(coupons)
	admin := &User{UserId: "e9a5b807-fa8b-455a-8dfa-b189d28310ff"}
	coupons, err = CreateCoupons(ctx, admin, 10)
	assert.Nil(err)
	assert.Len(coupons, 10)
	for _, coupon := range coupons {
		assert.False(coupon.OccupiedBy.Valid)
		assert.False(coupon.OccupiedAt.Valid)
	}
	coupons, err = ReadCoupons(ctx)
	assert.Nil(err)
	assert.Len(coupons, 11)
	user, err = FindUser(ctx, user.UserId)
	assert.Nil(err)
	assert.NotNil(user)
	assert.True(user.SubscribedAt.After(genesisStartedAt()))

	user3, err := createUser(ctx, "accessToken", bot.UuidNewV4().String(), "11100", "name", "http://localhost")
	assert.Nil(err)
	assert.NotNil(user3)
	coupon, err = Occupied(ctx, coupon.Code, user3)
	assert.NotNil(err)
	assert.Nil(coupon)
}
