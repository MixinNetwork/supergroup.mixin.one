package externals

import (
	"context"

	"github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

func AssetList(ctx context.Context, authorizationID, private, scope string) ([]*bot.Asset, error) {
	// FixMe
	if authorizationID == "" {
		return bot.AssetList(ctx, private)
	}
	mixin := config.AppConfig.Mixin
	requestID := bot.UuidNewV4().String()
	token, err := bot.SignOauthAccessToken(mixin.ClientId, authorizationID, private, "GET", "/assets", "", scope, requestID)
	if err != nil {
		return nil, err
	}
	list, err := bot.AssetListWithRequestID(ctx, token, requestID)
	if err != nil {
		return nil, parseError(ctx, err.(bot.Error))
	}
	return list, nil
}

func AssetShow(ctx context.Context, assetId, authorizationID, private, scope string) (*bot.Asset, error) {
	// FixMe
	if authorizationID == "" {
		return bot.AssetShow(ctx, assetId, private)
	}
	mixin := config.AppConfig.Mixin
	requestID := bot.UuidNewV4().String()
	token, err := bot.SignOauthAccessToken(mixin.ClientId, authorizationID, private, "GET", "/assets/"+assetId, "", scope, requestID)
	if err != nil {
		return nil, err
	}
	asset, err := bot.AssetShowWithRequestID(ctx, assetId, token, requestID)
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
