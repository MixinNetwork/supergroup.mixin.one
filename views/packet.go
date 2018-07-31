package views

import (
	"net/http"
	"time"

	number "github.com/MixinNetwork/go-number"
	"github.com/MixinNetwork/supergroup.mixin.one/models"
)

type PreparationView struct {
	Conversation struct {
		PariticipantsCount int `json:"participants_count"`
	} `json:"conversation"`
	Assets []AssetView `json:"assets"`
}

type AssetView struct {
	Type     string `json:"type"`
	AssetId  string `json:"asset_id"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	IconURL  string `json:"icon_url"`
	Balance  string `json:"balance"`
	PriceBTC string `json:"price_btc"`
	PriceUSD string `json:"price_usd"`
}

type ParticipantView struct {
	Type      string    `json:"type"`
	UserId    string    `json:"user_id"`
	FullName  string    `json:"full_name"`
	AvatarURL string    `json:"avatar_url"`
	Amount    string    `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

type PacketView struct {
	Type            string            `json:"type"`
	PacketId        string            `json:"packet_id"`
	User            UserView          `json:"user"`
	Asset           AssetView         `json:"asset"`
	Amount          string            `json:"amount"`
	Greeting        string            `json:"greeting"`
	TotalCount      int64             `json:"total_count"`
	RemainingCount  int64             `json:"remaining_count"`
	RemainingAmount string            `json:"remaining_amount"`
	OpenedCount     int64             `json:"opened_count"`
	OpenedAmount    string            `json:"opened_amount"`
	State           string            `json:"state"`
	Participants    []ParticipantView `json:"participants"`
}

func buildAssetView(asset *models.Asset) AssetView {
	return AssetView{
		Type:     "asset",
		AssetId:  asset.AssetId,
		Symbol:   asset.Symbol,
		Name:     asset.Name,
		IconURL:  asset.IconURL,
		Balance:  asset.Balance,
		PriceBTC: asset.PriceBTC,
		PriceUSD: asset.PriceUSD,
	}
}

func buildParticipantsView(participants []*models.Participant) []ParticipantView {
	participantsView := make([]ParticipantView, len(participants))
	for i, p := range participants {
		participantsView[i] = ParticipantView{
			Type:      "participant",
			UserId:    p.UserId,
			FullName:  p.FullName,
			AvatarURL: p.AvatarURL,
			Amount:    p.Amount,
			CreatedAt: p.CreatedAt,
		}
	}
	return participantsView
}

func RenderPacketPreparation(w http.ResponseWriter, r *http.Request, participantsCount int, assets []*models.Asset) {
	assetsView := make([]AssetView, len(assets))
	for i, a := range assets {
		assetsView[i] = buildAssetView(a)
	}
	prepareView := PreparationView{
		Assets: assetsView,
	}
	prepareView.Conversation.PariticipantsCount = participantsCount
	RenderDataResponse(w, r, prepareView)
}

func RenderPacket(w http.ResponseWriter, r *http.Request, packet *models.Packet) {
	packetView := PacketView{
		Type:            "packet",
		PacketId:        packet.PacketId,
		Asset:           buildAssetView(packet.Asset),
		User:            buildUserView(packet.User),
		Amount:          packet.Amount,
		Greeting:        packet.Greeting,
		TotalCount:      packet.TotalCount,
		RemainingCount:  packet.RemainingCount,
		RemainingAmount: packet.RemainingAmount,
		OpenedCount:     packet.TotalCount - packet.RemainingCount,
		OpenedAmount:    number.FromString(packet.Amount).Sub(number.FromString(packet.RemainingAmount)).Persist(),
		State:           packet.State,
		Participants:    buildParticipantsView(packet.Participants),
	}
	RenderDataResponse(w, r, packetView)
}
