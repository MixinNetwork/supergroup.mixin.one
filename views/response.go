package views

import (
	"net/http"

	bot "github.com/MixinNetwork/bot-api-go-client/v2"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

type ResponseView struct {
	Data  interface{} `json:"data,omitempty"`
	Error error       `json:"error,omitempty"`
	Prev  string      `json:"prev,omitempty"`
	Next  string      `json:"next,omitempty"`
}

func RenderDataResponse(w http.ResponseWriter, r *http.Request, view interface{}) {
	session.Render(r.Context()).JSON(w, http.StatusOK, ResponseView{Data: view})
}

func RenderErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	sessionError, ok := err.(session.Error)
	if !ok {
		botError, ok := err.(bot.Error)
		if ok {
			sessionError = session.Error{
				Status:      botError.Status,
				Code:        botError.Code,
				Description: botError.Description,
			}
		} else {
			sessionError = session.ServerError(r.Context(), err)
		}
	}
	if sessionError.Code == 10001 {
		sessionError.Code = 500
	}
	session.Render(r.Context()).JSON(w, sessionError.Status, ResponseView{Error: sessionError})
}

func RenderBlankResponse(w http.ResponseWriter, r *http.Request) {
	session.Render(r.Context()).JSON(w, http.StatusOK, ResponseView{})
}
