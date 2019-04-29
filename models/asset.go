package models

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	bot "github.com/MixinNetwork/bot-api-go-client"
	number "github.com/MixinNetwork/go-number"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

const assets_DDL = `
CREATE TABLE IF NOT EXISTS assets (
	asset_id         VARCHAR(36) PRIMARY KEY CHECK (asset_id ~* '^[0-9a-f-]{36,36}$'),
	symbol           VARCHAR(512) NOT NULL,
	name             VARCHAR(512) NOT NULL,
	icon_url         VARCHAR(1024) NOT NULL,
	price_btc        VARCHAR(128) NOT NULL,
	price_usd        VARCHAR(128) NOT NULL
);
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
	}
	if err := upsertAssets(ctx, assets); err != nil {
		return assets, session.TransactionError(ctx, err)
	}
	sort.Slice(assets, func(i, j int) bool {
		valuei := number.FromString(assets[i].PriceUSD).Mul(number.FromString(assets[i].Balance))
		valuej := number.FromString(assets[j].PriceUSD).Mul(number.FromString(assets[j].Balance))
		if valuei.Cmp(valuej) > 0 {
			return true
		}
		if valuei.Cmp(valuej) < 0 {
			return false
		}
		price := number.FromString(assets[i].PriceUSD).Cmp(number.FromString(assets[j].PriceUSD))
		if price > 0 {
			return true
		}
		if price < 0 {
			return false
		}
		return assets[i].Symbol < assets[j].Symbol
	})
	return assets, nil
}

func (current *User) ShowAsset(ctx context.Context, assetId string) (*Asset, error) {
	a, err := bot.AssetShow(ctx, assetId, current.AccessToken)
	if err != nil {
		return nil, err
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
	err = upsertAssets(ctx, []*Asset{asset})
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return asset, nil
}

func upsertAssets(ctx context.Context, assets []*Asset) error {
	var values bytes.Buffer
	for i, a := range assets {
		if i > 0 {
			values.WriteString(",")
		}
		values.WriteString(fmt.Sprintf("('%s','%s','%s','%s','%s','%s')", a.AssetId, a.Symbol, a.Name, a.IconURL, a.PriceBTC, a.PriceUSD))
	}
	query := fmt.Sprintf("INSERT INTO assets (%s) VALUES %s ON CONFLICT (asset_id) DO UPDATE SET (icon_url,price_btc,price_usd)=(EXCLUDED.icon_url, EXCLUDED.price_btc, EXCLUDED.price_usd)", strings.Join(assetsCols, ","), values.String())
	_, err := session.Database(ctx).ExecContext(ctx, query)
	return err
}

func readAsset(ctx context.Context, tx *sql.Tx, assetId string) (*Asset, error) {
	query := fmt.Sprintf("SELECT %s FROM assets WHERE asset_id=$1", strings.Join(assetsCols, ","))
	row := tx.QueryRowContext(ctx, query, assetId)
	return assetFromRow(row)
}

func assetFromRow(row durable.Row) (*Asset, error) {
	var a Asset
	err := row.Scan(&a.AssetId, &a.Symbol, &a.Name, &a.IconURL, &a.PriceBTC, &a.PriceUSD)
	return &a, err
}
