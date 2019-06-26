package models

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/lib/pq"
)

const coupons_DDL = `
CREATE TABLE IF NOT EXISTS coupons (
	coupon_id         VARCHAR(36) PRIMARY KEY CHECK (coupon_id ~* '^[0-9a-f-]{36,36}$'),
	code              VARCHAR(512) NOT NULL,
	occupied_by       VARCHAR(36),
	occupied_at       TIMESTAMP WITH TIME ZONE,
	created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS coupons_codex ON coupons(code);
CREATE INDEX IF NOT EXISTS coupons_occupiedx ON coupons(occupied_by);
`

type Coupon struct {
	CouponId   string
	Code       string
	OccupiedBy sql.NullString
	OccupiedAt pq.NullTime
	CreatedAt  time.Time
}

var couponColums = []string{"coupon_id", "code", "occupied_by", "occupied_at", "created_at"}

func (c *Coupon) values() []interface{} {
	return []interface{}{c.CouponId, c.Code, c.OccupiedBy, c.OccupiedAt, c.CreatedAt}
}

func couponFromRow(row durable.Row) (*Coupon, error) {
	var c Coupon
	err := row.Scan(&c.CouponId, &c.Code, &c.OccupiedBy, &c.OccupiedAt, &c.CreatedAt)
	return &c, err
}

func ReadCoupons(ctx context.Context) ([]*Coupon, error) {
	query := fmt.Sprintf("SELECT %s FROM coupons WHERE occupied_by IS NULL LIMIT 100", strings.Join(couponColums, ","))
	rows, err := session.Database(ctx).QueryContext(ctx, query)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	defer rows.Close()

	var coupons []*Coupon
	for rows.Next() {
		coupon, err := couponFromRow(rows)
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		}
		coupons = append(coupons, coupon)
	}
	return coupons, nil
}

func CreateCoupons(ctx context.Context, user *User) ([]*Coupon, error) {
	if !user.isAdmin() {
		return nil, session.ForbiddenError(ctx)
	}
	var coupons []*Coupon
	for i := 0; i < 50; i++ {
		coupon, err := CreateCoupon(ctx)
		if err != nil {
			session.TransactionError(ctx, err)
			continue
		}
		coupons = append(coupons, coupon)
	}
	return coupons, nil
}

func CreateCoupon(ctx context.Context) (*Coupon, error) {
	coupon := &Coupon{
		CouponId:  bot.UuidNewV4().String(),
		Code:      randomCode(),
		CreatedAt: time.Now(),
	}

	params, positions := compileTableQuery(couponColums)
	query := fmt.Sprintf("INSERT INTO coupons (%s) VALUES (%s)", params, positions)
	_, err := session.Database(ctx).ExecContext(ctx, query, coupon.values()...)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return coupon, nil
}

func Occupied(ctx context.Context, code string, user *User) (*Coupon, error) {
	var coupon *Coupon
	query := fmt.Sprintf("UPDATE coupons SET (occupied_by,occupied_at)=($1,$2) WHERE coupon_id=$3")
	err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		coupon, err = findCouponById(ctx, tx, code)
		if err != nil {
			return err
		}
		coupon.OccupiedBy = sql.NullString{String: user.UserId, Valid: true}
		coupon.OccupiedAt = pq.NullTime{Time: time.Now(), Valid: true}
		_, err = tx.ExecContext(ctx, query, coupon.OccupiedBy, coupon.OccupiedAt, coupon.CouponId)
		if err != nil {
			return err
		}
		user.State, user.SubscribedAt = PaymentStatePaid, time.Now()
		_, err = tx.ExecContext(ctx, "UPDATE users SET (state,subscribed_at)=($1,$2) WHERE user_id=$3", user.State, user.SubscribedAt, user.UserId)
		return err
	})
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return coupon, nil
}

func findCouponById(ctx context.Context, tx *sql.Tx, code string) (*Coupon, error) {
	query := fmt.Sprintf("SELECT %s FROM coupons WHERE code=$1", strings.Join(couponColums, ","))
	row := tx.QueryRowContext(ctx, query, code)
	coupon, err := couponFromRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return coupon, err
}

func randomCode() string {
	var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, 6)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
