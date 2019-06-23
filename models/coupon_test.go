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

	coupon, err := CreateCoupon(ctx)
	assert.Nil(err)
	assert.NotNil(coupon)
	coupons, err := ReadCoupons(ctx)
	assert.Nil(err)
	assert.Len(coupons, 1)
	coupon, err = CreateCoupon(ctx)
	assert.Nil(err)
	assert.NotNil(coupon)
	coupons, err = ReadCoupons(ctx)
	assert.Nil(err)
	assert.Len(coupons, 2)

	user, err := createUser(ctx, "accessToken", bot.UuidNewV4().String(), "1000", "name", "http://localhost")
	assert.Nil(err)
	assert.NotNil(user)
	coupon, err = Occupied(ctx, coupon.Code, user)
	assert.Nil(err)
	assert.NotNil(coupons)
	assert.Equal(user.UserId, coupon.OccupiedBy.String)
	coupons, err = CreateCoupons(ctx, user)
	assert.NotNil(err)
	assert.Nil(coupons)
	admin := &User{UserId: "e9a5b807-fa8b-455a-8dfa-b189d28310ff"}
	coupons, err = CreateCoupons(ctx, admin)
	assert.Nil(err)
	assert.Len(coupons, 50)
}
