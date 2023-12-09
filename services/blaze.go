package services

import (
	"net/http"
	"net/url"

	bot "github.com/MixinNetwork/bot-api-go-client/v2"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/gorilla/websocket"
)

func ConnectMixinBlaze(clientId, sessionId, sessionKey string) (*websocket.Conn, error) {
	token, err := bot.SignAuthenticationToken(clientId, sessionId, sessionKey, "GET", "/", "")
	if err != nil {
		return nil, err
	}
	header := make(http.Header)
	header.Add("Authorization", "Bearer "+token)
	cfg := config.AppConfig
	host := cfg.Service.BlazeRoot[cfg.Service.Retry%len(cfg.Service.BlazeRoot)]
	u := url.URL{Scheme: "wss", Host: host, Path: "/"}
	dialer := &websocket.Dialer{
		Subprotocols: []string{"Mixin-Blaze-1"},
	}
	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		cfg.Service.Retry++
		return nil, err
	}
	return conn, nil
}
