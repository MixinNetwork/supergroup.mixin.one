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
	buildCommit := os.Getenv("TEST_BUILD_COMMIT")
	if buildCommit == "" {
		buildCommit = "dev"
	}
	wantUserAgent := "mixin-zh-group/" + buildCommit
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != wantUserAgent {
			t.Errorf("User-Agent = %q, want %q", got, wantUserAgent)
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
