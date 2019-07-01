package views

import (
	"net/http"
	"time"

	"github.com/MixinNetwork/supergroup.mixin.one/models"
)

type CouponView struct {
	Type      string    `json:"type"`
	CouponId  string    `json:"coupon_id"`
	Code      string    `json:"code"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

func buildCoupon(coupon *models.Coupon) CouponView {
	return CouponView{
		Type:      "coupon",
		CouponId:  coupon.CouponId,
		Code:      coupon.Code,
		FullName:  coupon.FullName,
		CreatedAt: coupon.CreatedAt,
	}
}

func RenderCoupon(w http.ResponseWriter, r *http.Request, coupon *models.Coupon) {
	if coupon != nil {
		RenderDataResponse(w, r, buildCoupon(coupon))
	} else {
		RenderBlankResponse(w, r)
	}
}

func RenderCoupons(w http.ResponseWriter, r *http.Request, coupons []*models.Coupon) {
	views := make([]CouponView, len(coupons))
	for i, coupon := range coupons {
		views[i] = buildCoupon(coupon)
	}
	RenderDataResponse(w, r, views)
}
