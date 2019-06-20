package views

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MixinNetwork/supergroup.mixin.one/models"
)

type UserView struct {
	Type           string `json:"type"`
	UserId         string `json:"user_id"`
	IdentityNumber string `json:"identity_number"`
	FullName       string `json:"full_name"`
	AvatarURL      string `json:"avatar_url"`
	SubscribedAt   string `json:"subscribed_at"`
	Role           string `json:"role"`
}

type AccountView struct {
	UserView
	AuthenticationToken string `json:"authentication_token"`
	TraceId             string `json:"trace_id"`
	State               string `json:"state"`
}

func buildUserView(user *models.User) UserView {
	return UserView{
		Type:           "user",
		UserId:         user.UserId,
		IdentityNumber: fmt.Sprint(user.IdentityNumber),
		FullName:       user.GetFullName(),
		AvatarURL:      user.AvatarURL,
		SubscribedAt:   user.SubscribedAt.Format(time.RFC3339Nano),
		Role:           user.GetRole(),
	}
}

func RenderUsersView(w http.ResponseWriter, r *http.Request, users []*models.User) {
	userViews := make([]UserView, len(users))
	for i, user := range users {
		userViews[i] = buildUserView(user)
	}
	RenderDataResponse(w, r, userViews)
}

func RenderUserView(w http.ResponseWriter, r *http.Request, user *models.User) {
	RenderDataResponse(w, r, buildUserView(user))
}

func RenderAccount(w http.ResponseWriter, r *http.Request, user *models.User) {
	userView := AccountView{
		UserView:            buildUserView(user),
		AuthenticationToken: user.AuthenticationToken,
		TraceId:             user.TraceId,
		State:               user.State,
	}
	RenderDataResponse(w, r, userView)
}
