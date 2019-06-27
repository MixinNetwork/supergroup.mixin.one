package models

import (
	"bytes"
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
	user_id	          VARCHAR(36) NOT NULL CHECK (user_id ~* '^[0-9a-f-]{36,36}$'),
	occupied_by       VARCHAR(36),
	occupied_at       TIMESTAMP WITH TIME ZONE,
	created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS coupons_codex ON coupons(code);
CREATE INDEX IF NOT EXISTS coupons_occupiedx ON coupons(occupied_by);
CREATE INDEX IF NOT EXISTS coupons_userx ON coupons(user_id);
`

type Coupon struct {
	CouponId   string
	Code       string
	UserId     string
	OccupiedBy sql.NullString
	OccupiedAt pq.NullTime
	CreatedAt  time.Time

	FullName string
}

var couponColums = []string{"coupon_id", "code", "user_id", "occupied_by", "occupied_at", "created_at"}

func (c *Coupon) values() []interface{} {
	return []interface{}{c.CouponId, c.Code, c.UserId, c.OccupiedBy, c.OccupiedAt, c.CreatedAt}
}

func couponFromRow(row durable.Row) (*Coupon, error) {
	var c Coupon
	err := row.Scan(&c.CouponId, &c.Code, &c.UserId, &c.OccupiedBy, &c.OccupiedAt, &c.CreatedAt)
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

func CreateCoupons(ctx context.Context, user *User, quantity int) ([]*Coupon, error) {
	if !user.isAdmin() {
		return nil, session.ForbiddenError(ctx)
	}
	if quantity > 100 || quantity < 1 {
		quantity = 100
	}
	var coupons []*Coupon

	var values bytes.Buffer
	t := time.Now()
	for i := 0; i < quantity; i++ {
		coupon := &Coupon{
			CouponId:  bot.UuidNewV4().String(),
			Code:      randomCode(),
			UserId:    user.UserId,
			CreatedAt: t,
		}
		coupons = append(coupons, coupon)
		if i > 0 {
			values.WriteString(",")
		}
		values.WriteString(fmt.Sprintf("('%s', '%s', '%s', '%s')", coupon.CouponId, coupon.Code, coupon.UserId, string(pq.FormatTimestamp(coupon.CreatedAt))))
	}
	query := fmt.Sprintf("INSERT INTO coupons (coupon_id,code,user_id,created_at) VALUES %s", values.String())
	_, err := session.Database(ctx).ExecContext(ctx, query)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return coupons, nil
}

func (user *User) Coupons(ctx context.Context) ([]*Coupon, error) {
	var coupons []*Coupon
	err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		query := fmt.Sprintf("SELECT %s FROM coupons WHERE occupied_by IS NULL LIMIT 100", strings.Join(couponColums, ","))
		rows, err := tx.QueryContext(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			coupon, err := couponFromRow(rows)
			if err != nil {
				return err
			}
			if coupon.OccupiedBy.Valid {
				user, err := findUserById(ctx, tx, coupon.OccupiedBy.String)
				if err != nil {
					return err
				} else if user != nil {
					coupon.FullName = user.FullName
				}
			}
			coupons = append(coupons, coupon)
		}
		if len(coupons) != 0 {
			return nil
		}

		var values bytes.Buffer
		t := time.Now()
		for i := 0; i < 2; i++ {
			coupon := &Coupon{
				CouponId:  bot.UuidNewV4().String(),
				Code:      randomCode(),
				UserId:    user.UserId,
				CreatedAt: t,
			}
			coupons = append(coupons, coupon)
			if i > 0 {
				values.WriteString(",")
			}
			values.WriteString(fmt.Sprintf("('%s', '%s', '%s', '%s')", coupon.CouponId, coupon.Code, coupon.UserId, string(pq.FormatTimestamp(coupon.CreatedAt))))
		}
		query = fmt.Sprintf("INSERT INTO coupons (coupon_id,code,user_id,created_at) VALUES %s", values.String())
		_, err = tx.ExecContext(ctx, query)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if sessionErr, ok := err.(session.Error); ok {
			return nil, sessionErr
		}
		return nil, session.TransactionError(ctx, err)
	}
	return coupons, nil
}

func Occupied(ctx context.Context, code string, user *User) (*Coupon, error) {
	if user.State != PaymentStatePending {
		return nil, session.ForbiddenError(ctx)
	}
	var coupon *Coupon
	err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		coupon, err = findCouponByCode(ctx, tx, code)
		if err != nil {
			return err
		} else if coupon == nil {
			return nil
		}
		if coupon.OccupiedBy.Valid {
			return session.ForbiddenError(ctx)
		}
		coupon.OccupiedBy = sql.NullString{String: user.UserId, Valid: true}
		coupon.OccupiedAt = pq.NullTime{Time: time.Now(), Valid: true}
		query := fmt.Sprintf("UPDATE coupons SET (occupied_by,occupied_at)=($1,$2) WHERE coupon_id=$3")
		_, err = tx.ExecContext(ctx, query, coupon.OccupiedBy, coupon.OccupiedAt, coupon.CouponId)
		if err != nil {
			return err
		}
		return user.paymentInTx(ctx, tx, PayMethodCoupon)
	})
	if err != nil {
		if sessionErr, ok := err.(session.Error); ok {
			return nil, sessionErr
		}
		return nil, session.TransactionError(ctx, err)
	}
	return coupon, nil
}

func findCouponByCode(ctx context.Context, tx *sql.Tx, code string) (*Coupon, error) {
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
