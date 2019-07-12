package routes

import (
	"fmt"
	"net/http"

	"github.com/MixinNetwork/supergroup.mixin.one/plugin"
	"github.com/MixinNetwork/supergroup.mixin.one/views"
	"github.com/dimfeld/httptreemux"
)

func registerPluginRoutes(router *httptreemux.TreeMux) {
	for groupName, handler := range plugin.Handlers() {
		group := router.NewGroup("/" + groupName)
		for _, method := range []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"} {
			group.Handle(method, "/*", convertToTreemuxHandler(handler))
		}
	}

	registerShortcuts(router)
}

func convertToTreemuxHandler(handler http.Handler) httptreemux.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		handler.ServeHTTP(w, r)
	}
}

type shortcutsImpl struct{}

func registerShortcuts(router *httptreemux.TreeMux) {
	impl := shortcutsImpl{}

	router.GET("/shortcuts", impl.index)
	router.GET("/shortcuts/:group_id/:item_id/redirect", impl.redirect)
}

func (shortcutsImpl) index(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	views.RenderDataResponse(w, r, plugin.Shortcut.AllGroups())
}

func (shortcutsImpl) redirect(w http.ResponseWriter, r *http.Request, params map[string]string) {
	group := plugin.Shortcut.FindGroup(params["group_id"])
	if group == nil {
		views.RenderErrorResponse(w, r, fmt.Errorf("unknown shortcut group %s", params["group_id"]))
		return
	}

	item := group.FindItem(params["item_id"])
	if item == nil {
		views.RenderErrorResponse(w, r, fmt.Errorf("unknown shortcut item %s", params["item_id"]))
		return
	}

	http.Redirect(w, r, item.URL+"?token="+r.Header.Get("Authorization")[7:], 302)
}
