package externals

import (
	"context"
	"errors"

	"github.com/MixinNetwork/bot-api-go-client/v3"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

func AssetShow(ctx context.Context, assetId, authorizationID, private, scope string) (*bot.AssetNetwork, error) {
	// FixMe
	return bot.ReadAsset(ctx, assetId, nil)
}

func parseError(ctx context.Context, err error) error {
	if session.IsUnauthorized(err) {
		return session.AuthorizationError(ctx)
	}
	var apiError bot.Error
	if errors.As(err, &apiError) && apiError.Code > 0 {
		switch apiError.Code {
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
	return session.ServerError(ctx, err)
}
