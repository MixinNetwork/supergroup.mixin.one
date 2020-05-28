package externals

import (
	"context"

	"github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

func AssetList(ctx context.Context, token string) ([]bot.Asset, error) {
	list, err := bot.AssetList(ctx, token)
	if err != nil {
		return nil, parseError(ctx, err.(bot.Error))
	}
	return list, nil
}

func AssetShow(ctx context.Context, assetId, token string) (*bot.Asset, error) {
	asset, err := bot.AssetShow(ctx, assetId, token)
	if err != nil {
		return nil, parseError(ctx, err.(bot.Error))
	}
	return asset, nil
}

func parseError(ctx context.Context, err bot.Error) error {
	if err.Code > 0 {
		switch err.Code {
		case 401:
			return session.AuthorizationError(ctx)
		case 403:
			return session.ForbiddenError(ctx)
		case 404:
			return session.NotFoundError(ctx)
		default:
			return session.ServerError(ctx, err)
		}
	}
	return nil
}
