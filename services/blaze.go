package services

import (
	"net/http"
	"net/url"
	"time"

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
	host := cfg.Service.BlazeRoot[0]
	u := url.URL{Scheme: "wss", Host: host, Path: "/"}
	dialer := &websocket.Dialer{
		Subprotocols:     []string{"Mixin-Blaze-1"},
		HandshakeTimeout: time.Second * 45,
	}
	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
