package externals

import (
	"context"

	"github.com/MixinNetwork/bot-api-go-client/v2"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

func AssetShow(ctx context.Context, assetId, authorizationID, private, scope string) (*bot.Asset, error) {
	// FixMe
	return bot.ReadAsset(ctx, assetId)
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
