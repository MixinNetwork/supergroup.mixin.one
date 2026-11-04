package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MixinNetwork/supergroup.mixin.one/config"
)

func TestRequestUserAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("User-Agent"), config.UserAgent(); got != want {
			t.Errorf("User-Agent = %q, want %q", got, want)
		}
	}))
	defer server.Close()
	old := config.AppConfig
	config.AppConfig = &config.Config{}
	config.AppConfig.Service.APIRoot = []string{server.URL}
	t.Cleanup(func() { config.AppConfig = old })
	if _, err := request(context.Background(), t.Name(), http.MethodGet, "/", nil, ""); err != nil {
		t.Fatal(err)
	}
}
