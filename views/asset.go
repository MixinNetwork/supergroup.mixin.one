package views

import (
	"net/http"

	"github.com/MixinNetwork/supergroup.mixin.one/models"
)

func RenderAssets(w http.ResponseWriter, r *http.Request, assets []*models.Asset) {
	assetsView := make([]AssetView, len(assets))
	for i, a := range assets {
		assetsView[i] = buildAssetView(a)
	}
	RenderDataResponse(w, r, assetsView)
}
