package views

import (
	"net/http"
	"time"

	"github.com/MixinNetwork/supergroup.mixin.one/models"
)

type MessageView struct {
	Type      string    `json:"type"`
	MessageId string    `json:"message_id"`
	Category  string    `json:"category"`
	Data      string    `json:"data"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

func buildMessageView(message *models.Message) MessageView {
	view := MessageView{
		Type:      "message",
		MessageId: message.MessageId,
		Category:  message.Category,
		Data:      message.Data,
		FullName:  message.FullName.String,
		CreatedAt: message.CreatedAt,
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
