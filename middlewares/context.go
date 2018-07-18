package middlewares

import (
	"net/http"

	"cloud.google.com/go/spanner"
	"github.com/MixinMessenger/supergroup.mixin.one/durable"
	"github.com/MixinMessenger/supergroup.mixin.one/session"
	"github.com/unrolled/render"
)

func Context(handler http.Handler, spannerClient *spanner.Client, render *render.Render) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := durable.WrapDatabase(spannerClient, nil)
		ctx := session.WithRequest(r.Context(), r)
		ctx = session.WithDatabase(ctx, db)
		ctx = session.WithRender(ctx, render)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}
