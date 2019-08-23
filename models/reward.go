package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	number "github.com/MixinNetwork/go-number"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

const rewards_DDL = `
CREATE TABLE IF NOT EXISTS rewards (
	reward_id	          VARCHAR(36) PRIMARY KEY CHECK (reward_id ~* '^[0-9a-f-]{36,36}$'),
	user_id	            VARCHAR(36) NOT NULL CHECK (user_id ~* '^[0-9a-f-]{36,36}$'),
	recipient_id	      VARCHAR(36) NOT NULL CHECK (user_id ~* '^[0-9a-f-]{36,36}$'),
	asset_id            VARCHAR(36) NOT NULL CHECK (asset_id ~* '^[0-9a-f-]{36,36}$'),
	amount              VARCHAR(128) NOT NULL,
	paid_at             TIMESTAMP WITH TIME ZONE NOT NULL,
	created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
`

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
	return &r, err
}

func CreateReward(ctx context.Context, traceId, userId, recipientId, assetId, amount string) (*Reward, error) {
	var reward *Reward
	err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
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
		params, positions := compileTableQuery(rewardColumns)
		_, err = tx.ExecContext(ctx, fmt.Sprintf("INSERT INTO rewards (%s) VALUES (%s)", params, positions), reward.values()...)
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
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return reward, err
}
