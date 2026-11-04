package services

import (
	"net/http"
	"net/url"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client/v3"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/gorilla/websocket"
)

func ConnectMixinBlaze(clientId, sessionId, sessionKey string) (*websocket.Conn, error) {
	token, err := bot.SignAuthenticationToken("GET", "/", "", bot.NewSafeUser(clientId, sessionId, config.SDKPrivateKey(sessionKey)))
	if err != nil {
		return nil, err
	}
	header := make(http.Header)
	header.Add("Authorization", "Bearer "+token)
	header.Set("User-Agent", config.UserAgent())
	cfg := config.AppConfig
	host := cfg.Service.BlazeRoot[0]
	u := url.URL{Scheme: "wss", Host: host, Path: "/"}
	dialer := &websocket.Dialer{
		Subprotocols:     []string{"Mixin-Blaze-1"},
		HandshakeTimeout: time.Second * 45,
	}
	conn, resp, err := dialer.Dial(u.String(), header)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return nil, bot.AuthorizationError(nil)
		}
		return nil, err
	}
	return conn, nil
}
