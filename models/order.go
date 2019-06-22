package models

import (
	"context"
	"log"
	"strconv"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/lib/pq"
	"github.com/objcoding/wxpay"
)

const order_DDL = `
CREATE TABLE IF NOT EXISTS orders (
	order_id	      VARCHAR(36) PRIMARY KEY CHECK (order_id ~* '^[0-9a-f-]{36,36}$'),
	trace_id	      BIGSERIAL,
	user_id	          VARCHAR(36) NOT NULL CHECK (user_id ~* '^[0-9a-f-]{36,36}$'),
	prepay_id 	      VARCHAR(36) DEFAULT '',
	state             VARCHAR(32) NOT NULL,
	amount            VARCHAR(128) NOT NULL,
	channel           VARCHAR(32) NOT NULL,
	transaction_id    VARCHAR(32) DEFAULT '',
	qr_url    		  VARCHAR(64) DEFAULT '',
	created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	paid_at           TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS order_created_paidx ON orders(created_at, paid_at);
`

type Order struct {
	OrderId       string
	UserId        string
	TraceId       int64
	PrepayId      string
	State         string
	Amount        string
	Channel       string
	TransactionId string
	QrUrl         string
	CreatedAt     time.Time
	PaidAt        pq.NullTime
}

func GetNotPaidOrders(ctx context.Context) ([]*Order, error) {
	query := "SELECT * FROM orders WHERE state='NOTPAID' ORDER BY created_at"
	rows, err := session.Database(ctx).QueryContext(ctx, query)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}

	var orders []*Order
	for rows.Next() {
		p, err := orderFromRow(rows)
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		}
		orders = append(orders, p)
	}
	return orders, nil
}

func CreateOrder(ctx context.Context, userId, channel, amount string) (*Order, error) {
	order := &Order{
		OrderId:       bot.UuidNewV4().String(),
		UserId:        userId,
		TraceId:       0,
		PrepayId:      "",
		State:         "PENDING",
		Amount:        "0.0",
		Channel:       "wx",
		TransactionId: "",
		QrUrl:         "",
	}
	// create a record
	var err error
	query := "INSERT INTO orders (order_id, user_id, prepay_id, state, amount, channel) VALUES ($1, $2, $3, $4, $5, $6)"
	_, err = session.Database(ctx).ExecContext(ctx, query,
		order.OrderId, order.UserId, order.PrepayId, order.State, order.Amount, order.Channel)
	if err != nil {
		return nil, err
	}
	order, err = GetOrder(ctx, order.UserId, order.OrderId)
	// create wx payment request
	var wxp wxpay.Params
	wxp, err = createWxPayment(order.TraceId)
	if err != nil {
		return nil, err
	}
	log.Printf("%v, %v\n", wxp, err)
	order.QrUrl = wxp["code_url"]
	order.State = "NOTPAID"
	// update record
	query = "UPDATE orders SET (qr_url, state)=($1,$2) WHERE order_id=$3"
	_, err = session.Database(ctx).ExecContext(ctx, query, order.QrUrl, order.State, order.OrderId)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func GetOrder(ctx context.Context, userId, orderId string) (*Order, error) {
	query := "SELECT * FROM orders WHERE user_id=$1 and order_id=$2 ORDER BY created_at"
	rows, err := session.Database(ctx).QueryContext(ctx, query, userId, orderId)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	for rows.Next() {
		p, err := orderFromRow(rows)
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		}
		return p, nil
	}
	return nil, nil
}

func orderFromRow(row durable.Row) (*Order, error) {
	var od Order
	err := row.Scan(
		&od.OrderId,
		&od.TraceId,
		&od.UserId,
		&od.PrepayId,
		&od.State,
		&od.Amount,
		&od.Channel,
		&od.TransactionId,
		&od.QrUrl,
		&od.CreatedAt,
		&od.PaidAt,
	)
	return &od, err
}

func createWxClient() *wxpay.Client {
	cfg := config.Get()
	account := wxpay.NewAccount(cfg.Wechat.AppId, cfg.Wechat.MchId, cfg.Wechat.MchKey, false)
	client := wxpay.NewClient(account)
	// account.SetCertData("证书地址")
	client.SetAccount(account)
	client.SetHttpConnectTimeoutMs(2000)
	client.SetHttpReadTimeoutMs(1000)
	client.SetSignType(wxpay.HMACSHA256)
	return client
}

func createWxPayment(traceId int64) (wxpay.Params, error) {
	client := createWxClient()
	tradeNo := "tn-" + strconv.FormatInt(traceId, 10)
	// Place wx order
	params := make(wxpay.Params)
	params.SetString("body", "test").
		SetString("out_trade_no", tradeNo).
		SetInt64("total_fee", 1).
		SetString("spbill_create_ip", "127.0.0.1").
		SetString("notify_url", "https://xue.cn/").
		SetString("body", "学到-入群付费").
		SetString("trade_type", "NATIVE")
	p, err := client.UnifiedOrder(params)
	return p, err
}
