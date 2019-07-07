package routes

import (
	"net/http"

	"github.com/MixinNetwork/supergroup.mixin.one/plugin"
	"github.com/dimfeld/httptreemux"
)

func registerPluginRoutes(router *httptreemux.TreeMux) {
	for groupName, handler := range plugin.Handlers() {
		group := router.NewGroup("/" + groupName)
		for _, method := range []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"} {
			group.Handle(method, "/*", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				handler.ServeHTTP(w, r)
			})
		}
	}
}
