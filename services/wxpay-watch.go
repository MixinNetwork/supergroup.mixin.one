package services

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/objcoding/wxpay"

	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/models"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

type WxPaymentService struct{}

func StartWxPaymentWatch(name string, db *durable.Database) {
	context := session.WithDatabase(context.Background(), db)
	client := models.CreateWxClient()
	ctx := session.WithLogger(context, durable.BuildLogger())
	var orders []*models.Order
	var err error
	var params wxpay.Params
	for true {
		orders, err = models.GetNotPaidOrders(ctx)
		if err != nil {
			time.Sleep(time.Duration(7) * time.Second)
			continue
		}
		for _, order := range orders {
			params, err = models.FetchWxPayment(client, order.TraceId)
			if err != nil {
				time.Sleep(time.Duration(1) * time.Second)
				continue
			}
			if params["result_code"] == "SUCCESS" && params["trade_state"] == "SUCCESS" {
				tn := params["out_trade_no"]
				if strings.HasPrefix(tn, models.WX_TN_PREFIX) {
					if tnId, err := strconv.ParseInt(tn[3:], 10, 64); err == nil {
						log.Printf("tn:%v\n", tnId)
						models.UpdateOrderStateByTraceId(ctx, tnId, "PAID")
						if user, err := models.FindUser(ctx, order.UserId); err == nil {
							user.Payment(ctx)
						}
					}
				}
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
		time.Sleep(time.Duration(7) * time.Second)
	}
}
