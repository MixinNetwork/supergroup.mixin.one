package models

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

type Blacklist struct {
	UserId string
}

func (user *User) CreateBlacklist(ctx context.Context, userId string) (*Blacklist, error) {
	if id, _ := bot.UuidFromString(userId); id.String() != userId {
		return nil, session.BadDataError(ctx)
	}
	operators := config.AppConfig.System.Operators
	if !operators[user.UserId] || operators[userId] {
		return nil, nil
	}

	b := &Blacklist{UserId: userId}
	err := session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		u, err := findUserById(ctx, tx, userId)
		if err != nil || u == nil {
			return err
		}
		data := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("Banned %s, Mixin ID: %d", u.FullName, u.IdentityNumber)))
		err = createSystemDistributedMessage(ctx, tx, user, MessageCategoryPlainText, data)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO blacklists (user_id) VALUES ($1)", u.UserId)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "DELETE FROM users WHERE user_id=$1", u.UserId)
		return err
	})
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return b, nil
}

func ReadBlacklist(ctx context.Context, userId string) (*Blacklist, error) {
	var b *Blacklist
	err := session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		b, err = readBlacklistInTx(ctx, tx, userId)
		return err
	})
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return b, nil
}

func readBlacklistInTx(ctx context.Context, tx *sql.Tx, userId string) (*Blacklist, error) {
	var b Blacklist
	err := tx.QueryRowContext(ctx, "SELECT user_id from blacklists WHERE user_id=$1", userId).Scan(&b.UserId)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &b, nil
}
