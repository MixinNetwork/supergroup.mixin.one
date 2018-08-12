package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/MixinNetwork/supergroup.mixin.one/middlewares"
	"github.com/MixinNetwork/supergroup.mixin.one/models"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/MixinNetwork/supergroup.mixin.one/views"
	"github.com/dimfeld/httptreemux"
)

type usersImpl struct{}

type userRequest struct {
	FullName string `json:"full_name"`
}

func registerUsers(router *httptreemux.TreeMux) {
	impl := &usersImpl{}
	router.POST("/auth", impl.authenticate)
	router.POST("/account", impl.update)
	router.POST("/subscribe", impl.subscribe)
	router.POST("/unsubscribe", impl.unsubscribe)
	router.POST("/users/:id/remove", impl.remove)
	router.POST("/users/:id/block", impl.block)
	router.GET("/me", impl.me)
	router.GET("/subscribers", impl.subscribers)
	router.GET("/amount", impl.amount)
}

func (impl *usersImpl) authenticate(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if user, err := models.AuthenticateUserByOAuth(r.Context(), body.Code); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderAccount(w, r, user)
	}
}

func (impl *usersImpl) update(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body userRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
		return
	}
	current := middlewares.CurrentUser(r)
	if err := current.UpdateProfile(r.Context(), body.FullName); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderAccount(w, r, current)
	}
}

func (impl *usersImpl) me(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	views.RenderAccount(w, r, middlewares.CurrentUser(r))
}

func (impl *usersImpl) subscribers(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	offset, _ := time.Parse(time.RFC3339Nano, r.URL.Query().Get("offset"))
	if users, err := models.Subscribers(r.Context(), offset); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderUsersView(w, r, users)
	}
}

func (impl *usersImpl) amount(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	if amount, err := models.SubscribersCount(r.Context()); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderDataResponse(w, r, map[string]int64{"users_count": amount})
	}
}

func (impl *usersImpl) subscribe(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	current := middlewares.CurrentUser(r)
	if err := current.Subscribe(r.Context()); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderAccount(w, r, current)
	}
}

func (impl *usersImpl) unsubscribe(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	current := middlewares.CurrentUser(r)
	if err := current.Unsubscribe(r.Context()); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderAccount(w, r, current)
	}
}

func (impl *usersImpl) remove(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if err := middlewares.CurrentUser(r).DeleteUser(r.Context(), params["id"]); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderBlankResponse(w, r)
	}
}

func (impl *usersImpl) block(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if _, err := middlewares.CurrentUser(r).CreateBlacklist(r.Context(), params["id"]); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderBlankResponse(w, r)
	}
}
