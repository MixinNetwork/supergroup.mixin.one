package models

import (
	"context"
	"database/sql"
	"testing"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/stretchr/testify/assert"
)

func TestAssetCRUD(t *testing.T) {
	assert := assert.New(t)
	ctx := setupTestContext()
	defer teardownTestContext(ctx)

	assets := []*Asset{
		&Asset{
			AssetId:  bot.UuidNewV4().String(),
			Symbol:   "XIN",
			Name:     "Mixin",
			IconURL:  "http://mixin.one",
			PriceBTC: "0",
			PriceUSD: "0",
		},
	}
	err := upsertAssets(ctx, assets)
	assert.Nil(err)
	asset, err := testReadAsset(ctx, assets[0].AssetId)
	assert.Nil(err)
	assert.NotNil(asset)
	assert.Equal("XIN", asset.Symbol)
	assert.Equal("0", asset.PriceBTC)
	asset.PriceBTC = "0.1"
	assets = []*Asset{asset}
	err = upsertAssets(ctx, assets)
	assert.Nil(err)
	asset, err = testReadAsset(ctx, assets[0].AssetId)
	assert.Nil(err)
	assert.NotNil(asset)
	assert.Equal("0.1", asset.PriceBTC)
}

func testReadAsset(ctx context.Context, id string) (*Asset, error) {
	var asset *Asset
	err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		asset, err = readAsset(ctx, tx, id)
		return err
	})
	return asset, err
}
