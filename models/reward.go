package models

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"

	number "github.com/MixinNetwork/go-number"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/gofrs/uuid"
)

type Reward struct {
	RewardId    string
	UserId      string
	RecipientId string
	AssetId     string
	Amount      string
	PaidAt      time.Time
	CreatedAt   time.Time
}

var rewardColumns = []string{"reward_id", "user_id", "recipient_id", "asset_id", "amount", "paid_at", "created_at"}

func (r *Reward) values() []interface{} {
	return []interface{}{r.RewardId, r.UserId, r.RecipientId, r.AssetId, r.Amount, r.PaidAt, r.CreatedAt}
}

func rewardFromRow(row durable.Row) (*Reward, error) {
	var r Reward
	err := row.Scan(&r.RewardId, &r.UserId, &r.RecipientId, &r.AssetId, &r.Amount, &r.PaidAt, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &r, err
}

func CreateReward(ctx context.Context, traceId, userId, recipientId, assetId, amount string) (*Reward, error) {
	var reward *Reward
	err := session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		r, err := readRewardById(ctx, tx, traceId)
		if err != nil {
			return err
		}
		if r != nil {
			reward = r
			return nil
		}
		user, err := findUserById(ctx, tx, userId)
		if err != nil || user == nil {
			return err
		}
		recipient, err := findUserById(ctx, tx, recipientId)
		if err != nil || recipient == nil {
			return err
		}
		asset, err := findAssetById(ctx, tx, assetId)
		if err != nil || asset == nil {
			return err
		}
		if number.FromString(amount).Cmp(number.Zero()) <= 0 {
			return nil
		}

		reward = &Reward{
			RewardId:    traceId,
			UserId:      userId,
			RecipientId: recipientId,
			AssetId:     assetId,
			Amount:      amount,
			PaidAt:      time.Time{},
			CreatedAt:   time.Now(),
		}
		_, err = tx.ExecContext(ctx, durable.PrepareQuery("INSERT INTO rewards (%s) VALUES (%s)", rewardColumns), reward.values()...)
		if err != nil {
			return err
		}
		return createSystemRewardMessage(ctx, tx, reward, user, recipient, asset)
	})
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return reward, nil
}

func readRewardById(ctx context.Context, tx *sql.Tx, id string) (*Reward, error) {
	query := fmt.Sprintf("SELECT %s FROM rewards WHERE reward_id=$1", strings.Join(rewardColumns, ","))
	row := tx.QueryRowContext(ctx, query, id)
	reward, err := rewardFromRow(row)
	return reward, err
}

func PendingRewards(ctx context.Context, limit int) ([]*Reward, error) {
	query := fmt.Sprintf("SELECT %s FROM rewards WHERE paid_at=$1 LIMIT %d", strings.Join(rewardColumns, ","), limit)
	rows, err := session.Database(ctx).QueryContext(ctx, query, time.Time{})
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	defer rows.Close()

	var rewards []*Reward
	for rows.Next() {
		r, err := rewardFromRow(rows)
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		}
		rewards = append(rewards, r)
	}
	return rewards, nil
}

func SendRewardTransfer(ctx context.Context, reward *Reward) error {
	return nil
	/*
		TODO:
		traceId, err := generateRewardId(reward.RewardId)
		if err != nil {
			return session.ServerError(ctx, err)
		}
		if !reward.PaidAt.IsZero() {
			return nil
		}
		user, err := FindUser(ctx, reward.UserId)
		if err != nil {
			return err
		}
		memo := fmt.Sprintf(config.AppConfig.MessageTemplate.MessageRewardMemo, user.FullName)
		if len(memo) > 140 {
			memo = memo[:120]
		}
			in := &bot.TransferInput{
				AssetId:     reward.AssetId,
				RecipientId: reward.RecipientId,
				Amount:      number.FromString(reward.Amount),
				TraceId:     traceId,
				Memo:        memo,
			}
			mixin := config.AppConfig.Mixin
			_, err = bot.CreateTransfer(ctx, in, mixin.ClientId, mixin.SessionId, mixin.SessionKey, mixin.SessionAssetPIN, mixin.PinToken)
			if err != nil {
				return session.ServerError(ctx, err)
			}
			return UpdateReward(ctx, reward.RewardId)
	*/
}

func UpdateReward(ctx context.Context, rewardId string) error {
	query := "UPDATE rewards SET paid_at=$1 WHERE reward_id=$2"
	_, err := session.Database(ctx).ExecContext(ctx, query, time.Now(), rewardId)
	if err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func generateRewardId(rewardId string) (string, error) {
	h := md5.New()
	io.WriteString(h, rewardId)
	io.WriteString(h, "REWARD")
	sum := h.Sum(nil)
	sum[6] = (sum[6] & 0x0f) | 0x30
	sum[8] = (sum[8] & 0x3f) | 0x80
	id, err := uuid.FromBytes(sum)
	return id.String(), err
}
