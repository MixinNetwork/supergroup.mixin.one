package routes

import (
	"encoding/json"
	"net/http"

	number "github.com/MixinNetwork/go-number"
	"github.com/MixinNetwork/supergroup.mixin.one/middlewares"
	"github.com/MixinNetwork/supergroup.mixin.one/models"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/MixinNetwork/supergroup.mixin.one/views"
	"github.com/dimfeld/httptreemux"
)

type packetsImpl struct{}

type packetRequest struct {
	ConversationId string `json:"conversation_id"`
	AssetId        string `json:"asset_id"`
	Amount         string `json:"amount"`
	TotalCount     int64  `json:"total_count"`
	Greeting       string `json:"greeting"`
}

func registerPackets(router *httptreemux.TreeMux) {
	impl := &packetsImpl{}

	router.GET("/packets/prepare", impl.prepare)
	router.POST("/packets", impl.create)
	router.GET("/packets/:id", impl.show)
	router.POST("/packets/:id/claim", impl.claim)
}

func (impl *packetsImpl) prepare(w http.ResponseWriter, r *http.Request, params map[string]string) {
	current := middlewares.CurrentUser(r)
	if participantsCount, err := current.Prepare(r.Context()); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else if assets, err := current.ListAssets(r.Context()); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderPacketPreparation(w, r, participantsCount, assets)
	}
}

func (impl *packetsImpl) create(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body packetRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
	} else if packet, err := middlewares.CurrentUser(r).CreatePacket(r.Context(), body.AssetId, number.FromString(body.Amount), body.TotalCount, body.Greeting); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderPacket(w, r, packet)
	}
}

func (impl *packetsImpl) show(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if packet, err := models.ShowPacket(r.Context(), params["id"]); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else if packet == nil {
		views.RenderErrorResponse(w, r, session.NotFoundError(r.Context()))
	} else {
		views.RenderPacket(w, r, packet)
	}
}

func (impl *packetsImpl) claim(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if packet, err := middlewares.CurrentUser(r).ClaimPacket(r.Context(), params["id"]); err != nil {
		views.RenderErrorResponse(w, r, err)
	} else if packet == nil {
		views.RenderErrorResponse(w, r, session.NotFoundError(r.Context()))
	} else {
		views.RenderPacket(w, r, packet)
	}
}
