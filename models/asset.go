package models

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/externals"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

const (
	PacketMinAmount = "0.0001"
)

type Asset struct {
	AssetId  string
	Symbol   string
	Name     string
	IconURL  string
	PriceBTC string
	PriceUSD string
}

var assetsColumns = []string{"asset_id", "symbol", "name", "icon_url", "price_btc", "price_usd"}

func (a *Asset) values() []interface{} {
	return []interface{}{a.AssetId, a.Symbol, a.Name, a.IconURL, a.PriceBTC, a.PriceUSD}
}

func assetFromRow(row durable.Row) (*Asset, error) {
	var a Asset
	err := row.Scan(&a.AssetId, &a.Symbol, &a.Name, &a.IconURL, &a.PriceBTC, &a.PriceUSD)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &a, err
}

func (current *User) ListAssets(ctx context.Context) ([]*Asset, error) {
	return nil, nil
}

func (current *User) ShowAsset(ctx context.Context, assetId string) (*Asset, error) {
	a, err := externals.AssetShow(ctx, assetId, current.AuthorizationID, current.AccessToken, current.Scope)
	if err != nil {
		return nil, err
	}
	asset := &Asset{
		AssetId:  a.AssetID,
		Symbol:   a.Symbol,
		Name:     a.Name,
		IconURL:  a.IconURL,
		PriceBTC: a.PriceBTC,
		PriceUSD: a.PriceUSD,
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
	query := fmt.Sprintf("INSERT INTO assets (%s) VALUES %s ON CONFLICT (asset_id) DO UPDATE SET (icon_url,price_btc,price_usd)=(EXCLUDED.icon_url, EXCLUDED.price_btc, EXCLUDED.price_usd)", strings.Join(assetsColumns, ","), values.String())
	_, err := session.Database(ctx).ExecContext(ctx, query)
	return err
}

func findAssetById(ctx context.Context, tx *sql.Tx, assetId string) (*Asset, error) {
	query := fmt.Sprintf("SELECT %s FROM assets WHERE asset_id=$1", strings.Join(assetsColumns, ","))
	row := tx.QueryRowContext(ctx, query, assetId)
	asset, err := assetFromRow(row)
	return asset, err
}
