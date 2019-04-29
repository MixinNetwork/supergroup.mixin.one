package models

import (
	"context"
	"database/sql"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

const blacklist_DDL = `
CREATE TABLE IF NOT EXISTS blacklists (
	user_id	          VARCHAR(36) PRIMARY KEY CHECK (user_id ~* '^[0-9a-f-]{36,36}$')
);
`

type Blacklist struct {
	UserId string
}

func (user *User) CreateBlacklist(ctx context.Context, userId string) (*Blacklist, error) {
	_, err := bot.UuidFromString(userId)
	if err != nil {
		return nil, session.ForbiddenError(ctx)
	}
	if !config.Operators[user.UserId] {
		return nil, nil
	}
	if config.Operators[userId] {
		return nil, nil
	}
	user, err = FindUser(ctx, userId)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	} else if user == nil {
		return nil, nil
	}

	err = session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO blacklists (user_id) VALUES ($1)", user.UserId)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "DELETE FROM users WHERE user_id=$1", user.UserId)
		return err
	})
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return &Blacklist{UserId: userId}, nil
}

func readBlacklist(ctx context.Context, userId string) (*Blacklist, error) {
	var b Blacklist
	err := session.Database(ctx).QueryRowContext(ctx, "SELECT user_id from blacklists WHERE user_id=$1", userId).Scan(&b.UserId)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	return &b, nil
}
