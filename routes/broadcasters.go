package routes

import (
	"encoding/json"
	"net/http"

	"github.com/MixinNetwork/supergroup.mixin.one/middlewares"
	"github.com/MixinNetwork/supergroup.mixin.one/models"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/MixinNetwork/supergroup.mixin.one/views"
	"github.com/dimfeld/httptreemux"
)

type broadcastersImpl struct{}

type broadcastRequest struct {
	Identity int64 `json:"identity"`
}

func registerBroadcasters(router *httptreemux.TreeMux) {
	impl := &broadcastersImpl{}

	router.POST("/broadcasters", impl.create)
	router.GET("/broadcasters", impl.index)
}

func (impl *broadcastersImpl) create(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	var body broadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	}

	user, err := middlewares.CurrentUser(r).CreateBroadcaster(r.Context(), body.Identity)
	if err != nil {
		views.RenderErrorResponse(w, r, err)
	} else if user == nil {
		views.RenderErrorResponse(w, r, session.NotFoundError(r.Context()))
	} else {
		views.RenderUserView(w, r, user)
	}
}

func (impl *broadcastersImpl) index(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	users, err := models.ReadBroadcasters(r.Context())
	if err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderUsersView(w, r, users)
	}
}
