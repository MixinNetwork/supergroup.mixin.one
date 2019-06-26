package routes

import (
	"net/http"

	"github.com/MixinNetwork/supergroup.mixin.one/middlewares"
	"github.com/MixinNetwork/supergroup.mixin.one/models"
	"github.com/MixinNetwork/supergroup.mixin.one/views"
	"github.com/dimfeld/httptreemux"
)

type couponImpl struct{}

func registerCoupons(router *httptreemux.TreeMux) {
	impl := &couponImpl{}

	router.POST("/coupons", impl.create)
	router.POST("/coupons/:code", impl.occupy)
	router.GET("/coupons", impl.index)
}

func (impl *couponImpl) create(w http.ResponseWriter, r *http.Request, params map[string]string) {
	coupons, err := models.CreateCoupons(r.Context(), middlewares.CurrentUser(r))
	if err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderCoupons(w, r, coupons)
	}
}

func (impl *couponImpl) occupy(w http.ResponseWriter, r *http.Request, params map[string]string) {
	coupon, err := models.Occupied(r.Context(), params["code"], middlewares.CurrentUser(r))
	if err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderCoupon(w, r, coupon)
	}
}

func (impl *couponImpl) index(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	coupons, err := models.ReadCoupons(r.Context())
	if err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderCoupons(w, r, coupons)
	}
}
