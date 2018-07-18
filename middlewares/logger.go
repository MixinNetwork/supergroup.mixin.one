package middlewares

import (
	"net/http"
	"strings"

	bot "github.com/MixinMessenger/bot-api-go-client"
	"github.com/MixinMessenger/supergroup.mixin.one/durable"
	"github.com/MixinMessenger/supergroup.mixin.one/session"
)

func Log(handler http.Handler, client *durable.LoggerClient, service string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.ToUpper(bot.NewV4().String())
		r.Header["X-Request-Id"] = []string{id}
		logger := durable.BuildLogger()
		ctx := session.WithLogger(r.Context(), logger)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}
