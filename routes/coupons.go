package routes

import (
	"encoding/json"
	"net/http"

	"github.com/MixinNetwork/supergroup.mixin.one/middlewares"
	"github.com/MixinNetwork/supergroup.mixin.one/models"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/MixinNetwork/supergroup.mixin.one/views"
	"github.com/dimfeld/httptreemux"
)

type couponImpl struct{}

type couponRequest struct {
	Quantity int `json:"quantity"`
}

func registerCoupons(router *httptreemux.TreeMux) {
	impl := &couponImpl{}

	router.POST("/coupons", impl.create)
	router.POST("/coupons/:code", impl.occupy)
	//router.GET("/coupon", impl.reward)
}

func (impl *couponImpl) create(w http.ResponseWriter, r *http.Request, params map[string]string) {
	var body couponRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		views.RenderErrorResponse(w, r, session.BadRequestError(r.Context()))
		return
	}
	coupons, err := models.CreateCoupons(r.Context(), middlewares.CurrentUser(r), body.Quantity)
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
	} else if coupon == nil {
		views.RenderErrorResponse(w, r, session.NotFoundError(r.Context()))
	} else {
		views.RenderCoupon(w, r, coupon)
	}
}

func (impl *couponImpl) reward(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	coupons, err := middlewares.CurrentUser(r).Coupons(r.Context())
	if err != nil {
		views.RenderErrorResponse(w, r, err)
	} else {
		views.RenderCoupons(w, r, coupons)
	}
}
