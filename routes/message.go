package routes

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/MixinNetwork/supergroup.mixin.one/middlewares"
	"github.com/MixinNetwork/supergroup.mixin.one/models"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/MixinNetwork/supergroup.mixin.one/views"
	"github.com/dimfeld/httptreemux"
)

type messageImpl struct{}

func registerMesseages(router *httptreemux.TreeMux) {
	impl := messageImpl{}

	router.GET("/messages", impl.index)
	router.GET("/messages/:id/recall", impl.recall)
}

func (impl *messageImpl) index(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	user := middlewares.CurrentUser(r)
	if user.GetRole() != "admin" {
		views.RenderErrorResponse(w, r, session.ForbiddenError(r.Context()))
	} else if messages, err := models.LastestMessageWithUser(r.Context(), 200); err != nil {
		views.RenderErrorResponse(w, r, session.ForbiddenError(r.Context()))
	} else {
		views.RenderMessages(w, r, messages)
	}
}

func (impl *messageImpl) recall(w http.ResponseWriter, r *http.Request, params map[string]string) {
	data, err := json.Marshal(map[string]string{"message_id": params["id"]})
	if err != nil {
		views.RenderErrorResponse(w, r, err)
	}
	str := base64.StdEncoding.EncodeToString(data)
	t := time.Now()
	_, err = models.CreateMessage(r.Context(), middlewares.CurrentUser(r), params["id"], models.MessageCategoryMessageRecall, "", str, t, t)
	if err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderBlankResponse(w, r)
	}
}
