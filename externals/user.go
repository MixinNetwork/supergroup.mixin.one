package externals

import (
	"context"
	"strings"

	"github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

func UserMeFromCode(ctx context.Context, code, private, public string) (*bot.User, string, string, error) {
	mixin := config.AppConfig.Mixin
	_, scope, authorizationID, err := bot.OAuthGetAccessToken(ctx, mixin.ClientId, mixin.ClientSecret, code, "", public)
	if err != nil {
		return nil, "", "", parseError(ctx, err.(bot.Error))
	}
	if !strings.Contains(scope, "PROFILE:READ") {
		return nil, "", "", session.ForbiddenError(ctx)
	}
	requestID := bot.UuidNewV4().String()
	token, err := bot.SignOauthAccessToken(mixin.ClientId, authorizationID, private, "GET", "/me", "", scope, requestID)
	if err != nil {
		return nil, "", "", err
	}
	me, err := bot.UserMeWithRequestID(ctx, token, requestID)
	if err != nil {
		return nil, "", "", parseError(ctx, err.(bot.Error))
	}
	return me, authorizationID, scope, nil
}

func UserMe(ctx context.Context, authorizationID, private, scope string) (*bot.User, error) {
	// FixMe
	if authorizationID == "" {
		return bot.UserMe(ctx, private)
	}
	mixin := config.AppConfig.Mixin
	requestID := bot.UuidNewV4().String()
	token, err := bot.SignOauthAccessToken(mixin.ClientId, authorizationID, private, "GET", "/me", "", scope, requestID)
	if err != nil {
		return nil, err
	}
	me, err := bot.UserMeWithRequestID(ctx, token, requestID)
	if err != nil {
		return nil, err
	}
	return me, nil
}
