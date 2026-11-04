package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client/v3"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

func TestRequestUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()
	old := config.AppConfig
	config.AppConfig = &config.Config{}
	config.AppConfig.Service.APIRoot = []string{server.URL}
	t.Cleanup(func() { config.AppConfig = old })
	_, err := request(context.Background(), t.Name(), http.MethodGet, "/", nil, "")
	if !session.IsUnauthorized(err) {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestBlazeDoesNotRetryUnauthorized(t *testing.T) {
	mc := &MessageContext{Transactions: newTmap(), WriteBuffer: make(chan []byte, 2)}
	timer := time.NewTimer(time.Minute)
	defer timer.Stop()
	drained := false
	done := make(chan error, 1)
	go func() { done <- writeMessageAndWait(context.Background(), mc, "TEST", nil, timer, &drained) }()
	var request BlazeMessage
	if err := json.Unmarshal(<-mc.WriteBuffer, &request); err != nil {
		t.Fatal(err)
	}
	handler := mc.Transactions.retrive(request.Id)
	if handler == nil {
		t.Fatal("missing transaction handler")
	}
	if err := handler(BlazeMessage{Error: &session.Error{Code: 401}}); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if !session.IsUnauthorized(err) {
			t.Fatalf("expected unauthorized error, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("unauthorized request did not return")
	}
	select {
	case <-mc.WriteBuffer:
		t.Fatal("unauthorized request was retried")
	default:
	}
}

func TestWrappedUnauthorized(t *testing.T) {
	err := session.ServerError(nil, bot.Error{Code: 401})
	if !session.IsUnauthorized(err) {
		t.Fatal("wrapped unauthorized error was not recognized")
	}
	if session.IsUnauthorized(session.ServerError(nil, bot.Error{Code: 500})) {
		t.Fatal("server error treated as unauthorized")
	}
	if !session.IsUnauthorized(session.ServerError(nil, session.Error{Code: 401})) {
		t.Fatal("nested session authorization error was not recognized")
	}
}
