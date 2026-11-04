package externals

import (
	"context"

	"github.com/MixinNetwork/bot-api-go-client/v3"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
)

func CreateConversation(ctx context.Context, category, participantId string) error {
	if config.AppConfig.Service.Environment == "test" {
		return nil
	}
	mixin := config.AppConfig.Mixin
	_, err := bot.CreateContactConversation(ctx, participantId, bot.NewSafeUser(mixin.ClientId, mixin.SessionId, config.SDKPrivateKey(mixin.SessionKey)))
	if err != nil {
		return parseError(ctx, err)
	}
	return nil
}

func ReadConversation(ctx context.Context, conversationID string) (*bot.Conversation, error) {
	mixin := config.AppConfig.Mixin
	token, err := bot.SignAuthenticationToken("GET", "/conversations/"+conversationID, "", bot.NewSafeUser(mixin.ClientId, mixin.SessionId, config.SDKPrivateKey(mixin.SessionKey)))
	if err != nil {
		return nil, err
	}
	return bot.ConversationShowByToken(ctx, conversationID, token)
}
