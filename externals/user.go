package externals

import (
	"context"
	"strings"

	"github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

func UserMe(ctx context.Context, code string) (*bot.User, string, error) {
	mixin := config.AppConfig.Mixin
	token, scope, err := bot.OAuthGetAccessToken(ctx, mixin.ClientId, mixin.ClientSecret, code, "")
	if err != nil {
		return nil, "", parseError(ctx, err.(bot.Error))
	}
	if !strings.Contains(scope, "PROFILE:READ") {
		return nil, "", session.ForbiddenError(ctx)
	}
	me, err := bot.UserMe(ctx, token)
	if err != nil {
		return nil, "", parseError(ctx, err.(bot.Error))
	}
	return me, token, nil
}
