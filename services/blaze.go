package services

import (
	"net/http"
	"net/url"

	bot "github.com/MixinMessenger/bot-api-go-client"
	"github.com/gorilla/websocket"
)

func ConnectMixinBlaze(clientId, sessionId, sessionKey string) (*websocket.Conn, error) {
	token, err := bot.SignAuthenticationToken(clientId, sessionId, sessionKey, "GET", "/", "")
	if err != nil {
		return nil, err
	}
	header := make(http.Header)
	header.Add("Authorization", "Bearer "+token)
	u := url.URL{Scheme: "wss", Host: "blaze.mixin.one", Path: "/"}
	dialer := &websocket.Dialer{
		Subprotocols: []string{"Mixin-Blaze-1"},
	}
	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
