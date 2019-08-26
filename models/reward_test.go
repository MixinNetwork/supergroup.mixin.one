package models

import (
	"testing"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/stretchr/testify/assert"
)

func TestRewardCRUD(t *testing.T) {
	assert := assert.New(t)
	ctx := setupTestContext()
	defer teardownTestContext(ctx)

	user, err := createUser(ctx, "accessToken", bot.UuidNewV4().String(), "1000", "name", "http://localhost")
	assert.Nil(err)
	assert.NotNil(user)
	recipient, err := createUser(ctx, "accessToken", bot.UuidNewV4().String(), "1100", "name", "http://localhost")
	assert.Nil(err)
	assert.NotNil(recipient)

	asset := &Asset{
		AssetId:  bot.UuidNewV4().String(),
		Symbol:   "XIN",
		Name:     "Mixin",
		IconURL:  "http://mixin.one",
		PriceBTC: "0",
		PriceUSD: "0",
	}
	err = upsertAssets(ctx, []*Asset{asset})
	assert.Nil(err)

	reward, err := CreateReward(ctx, bot.UuidNewV4().String(), user.UserId, recipient.UserId, asset.AssetId, "1")
	assert.Nil(err)
	assert.NotNil(reward)

	rewards, err := PendingRewards(ctx, 10)
	assert.Nil(err)
	assert.Len(rewards, 1)
	err = UpdateReward(ctx, reward.RewardId)
	assert.Nil(err)
	rewards, err = PendingRewards(ctx, 10)
	assert.Nil(err)
	assert.Len(rewards, 0)
}
