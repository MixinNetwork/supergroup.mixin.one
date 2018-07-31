package models

import (
	"context"
	"sort"

	"cloud.google.com/go/spanner"
	bot "github.com/MixinNetwork/bot-api-go-client"
	number "github.com/MixinNetwork/go-number"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"google.golang.org/api/iterator"
)

const assets_DDL = `
CREATE TABLE assets (
	asset_id         STRING(36) NOT NULL,
	symbol           STRING(512) NOT NULL,
	name             STRING(512) NOT NULL,
	icon_url         STRING(1024) NOT NULL,
	price_btc        STRING(128) NOT NULL,
	price_usd        STRING(128) NOT NULL,
) PRIMARY KEY(asset_id);
`

var assetsCols = []string{"asset_id", "symbol", "name", "icon_url", "price_btc", "price_usd"}

type Asset struct {
	AssetId  string
	Symbol   string
	Name     string
	IconURL  string
	PriceBTC string
	PriceUSD string
	Balance  string
}

func (current *User) ListAssets(ctx context.Context) ([]*Asset, error) {
	list, err := bot.AssetList(ctx, current.AccessToken)
	if err != nil {
		return nil, err
	}
	var assets []*Asset
	var mutations []*spanner.Mutation
	for _, a := range list {
		if number.FromString(a.Balance).Cmp(number.FromString("0.0001")) < 0 {
			continue
		}
		assets = append(assets, &Asset{
			AssetId:  a.AssetId,
			Symbol:   a.Symbol,
			Name:     a.Name,
			IconURL:  a.IconURL,
			PriceBTC: a.PriceBTC,
			PriceUSD: a.PriceUSD,
			Balance:  a.Balance,
		})
		vals := []interface{}{a.AssetId, a.Symbol, a.Name, a.IconURL, a.PriceBTC, a.PriceUSD}
		mutations = append(mutations, spanner.InsertOrUpdate("assets", assetsCols, vals))
	}
	if len(mutations) > 0 {
		err = session.Database(ctx).Apply(ctx, mutations, "assets", "INSERT", "ListAssets")
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		}
	}
	sort.Slice(assets, func(i, j int) bool {
		if number.FromString(assets[i].PriceUSD).Cmp(number.FromString(assets[j].PriceUSD)) > 0 {
			return true
		}
		if number.FromString(assets[i].PriceUSD).Cmp(number.FromString(assets[j].PriceUSD)) < 0 {
			return false
		}
		return assets[i].Symbol < assets[j].Symbol
	})
	return assets, nil
}

func (current *User) ShowAsset(ctx context.Context, assetId string) (*Asset, error) {
	a, err := bot.AssetShow(ctx, assetId, current.AccessToken)
	if err != nil {
		return nil, session.ServerError(ctx, err)
	}
	asset := &Asset{
		AssetId:  a.AssetId,
		Symbol:   a.Symbol,
		Name:     a.Name,
		IconURL:  a.IconURL,
		PriceBTC: a.PriceBTC,
		PriceUSD: a.PriceUSD,
		Balance:  a.Balance,
	}
	vals := []interface{}{a.AssetId, a.Symbol, a.Name, a.IconURL, a.PriceBTC, a.PriceUSD}
	err = session.Database(ctx).Apply(ctx, []*spanner.Mutation{
		spanner.InsertOrUpdate("assets", assetsCols, vals),
	}, "assets", "INSERT", "ShowAsset")
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return asset, nil
}

func readAsset(ctx context.Context, txn durable.Transaction, assetId string) (*Asset, error) {
	it := txn.Read(ctx, "assets", spanner.Key{assetId}, assetsCols)
	defer it.Stop()

	row, err := it.Next()
	if err == iterator.Done {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return assetFromRow(row)
}

func assetFromRow(row *spanner.Row) (*Asset, error) {
	var a Asset
	err := row.Columns(&a.AssetId, &a.Symbol, &a.Name, &a.IconURL, &a.PriceBTC, &a.PriceUSD)
	return &a, err
}
