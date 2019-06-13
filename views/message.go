package views

import (
	"net/http"

	"github.com/MixinNetwork/supergroup.mixin.one/models"
)

type MessageView struct {
	Type     string `json:"type"`
	Category string `json:"category"`
	Data     string `json:"data"`
	FullName string `json:"full_name"`
}

func buildMessageView(message *models.Message) MessageView {
	view := MessageView{
		Type:     "message",
		Category: message.Category,
		Data:     message.Data,
		FullName: message.FullName.String,
	}
	if view.FullName == "" {
		view.FullName = "NULL"
	}
	return view
}

func RenderMessages(w http.ResponseWriter, r *http.Request, messages []*models.Message) {
	views := make([]MessageView, len(messages))
	for i, message := range messages {
		views[i] = buildMessageView(message)
	}
	RenderDataResponse(w, r, views)
}
