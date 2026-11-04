package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	bot "github.com/MixinNetwork/bot-api-go-client/v3"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
)

func TestSDKUserAgent(t *testing.T) {
	if want := os.Getenv("TEST_BUILD_VERSION"); want != "" && config.BuildVersion != want {
		t.Fatalf("build version = %q, want %q", config.BuildVersion, want)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("User-Agent"), config.UserAgent(); got != want {
			t.Errorf("User-Agent = %q, want %q", got, want)
		}
	}))
	defer server.Close()
	bot.SetBaseUri(server.URL)
	t.Cleanup(func() { bot.SetBaseUri(bot.DefaultApiHost) })
	configureUserAgent()
	if _, err := bot.Request(context.Background(), http.MethodGet, "/", nil, ""); err != nil {
		t.Fatal(err)
	}
}
