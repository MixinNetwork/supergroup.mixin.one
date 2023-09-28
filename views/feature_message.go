package views

import (
	"net/http"
	"time"

	"github.com/MixinNetwork/supergroup.mixin.one/models"
)

type FeaturedMessageView struct {
	Type      string    `json:"type"`
	MessageId string    `json:"message_id"`
	Category  string    `json:"category"`
	Data      string    `json:"data"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

func buildFeaturedMessageView(message *models.FeaturedMessage) FeaturedMessageView {
	view := FeaturedMessageView{
		Type:      "featured_message",
		MessageId: message.MessageId,
		Category:  message.Category,
		Data:      message.Data,
		FullName:  message.FullName,
		CreatedAt: message.CreatedAt,
	}
	if view.FullName == "" {
		view.FullName = "NULL"
	}
	return view
}

func RenderFeaturedMessage(w http.ResponseWriter, r *http.Request, message *models.FeaturedMessage) {
	RenderDataResponse(w, r, buildFeaturedMessageView(message))
}

func RenderFeaturedMessages(w http.ResponseWriter, r *http.Request, messages []*models.FeaturedMessage) {
	views := make([]FeaturedMessageView, len(messages))
	for i, message := range messages {
		views[i] = buildFeaturedMessageView(message)
	}
	RenderDataResponse(w, r, views)
}
