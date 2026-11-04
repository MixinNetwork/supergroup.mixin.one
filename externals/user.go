package externals

import (
	"context"
	"strings"

	"github.com/MixinNetwork/bot-api-go-client/v3"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

func UserMeFromCode(ctx context.Context, code, private, public string) (*bot.UserMeView, string, string, error) {
	mixin := config.AppConfig.Mixin
	_, scope, authorizationID, err := bot.OAuthGetAccessToken(ctx, mixin.ClientId, mixin.ClientSecret, code, "", public, nil)
	if err != nil {
		return nil, "", "", parseError(ctx, err)
	}
	if !strings.Contains(scope, "PROFILE:READ") {
		return nil, "", "", session.ForbiddenError(ctx)
	}
	requestID := bot.UuidNewV4().String()
	token, err := bot.SignOauthAccessToken(mixin.ClientId, authorizationID, config.SDKPrivateKey(private), "GET", "/safe/me", "", scope, requestID)
	if err != nil {
		return nil, "", "", err
	}
	me, err := bot.UserMeWithRequestID(ctx, token, requestID)
	if err != nil {
		return nil, "", "", parseError(ctx, err)
	}
	return me, authorizationID, scope, nil
}

func UserMe(ctx context.Context, authorizationID, private, scope string) (*bot.UserMeView, error) {
	// FixMe
	if authorizationID == "" {
		return bot.UserMe(ctx, private)
	}
	mixin := config.AppConfig.Mixin
	requestID := bot.UuidNewV4().String()
	token, err := bot.SignOauthAccessToken(mixin.ClientId, authorizationID, config.SDKPrivateKey(private), "GET", "/safe/me", "", scope, requestID)
	if err != nil {
		return nil, err
	}
	me, err := bot.UserMeWithRequestID(ctx, token, requestID)
	if err != nil {
		return nil, err
	}
	return me, nil
}
