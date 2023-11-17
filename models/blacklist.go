package models

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"

	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/gofrs/uuid/v5"
	"github.com/lib/pq"
)

type Blacklist struct {
	UserId string
}

func (user *User) CreateBlacklist(ctx context.Context, userId string) (*Blacklist, error) {
	if id := uuid.FromStringOrNil(userId); id.String() != userId {
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
		data := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("Banned %s, Mixin ID: %d", u.FullName, u.IdentityNumber)))
		err = createSystemDistributedMessageInTx(ctx, tx, user, MessageCategoryPlainText, data)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "DELETE FROM users WHERE user_id=$1", u.UserId)
		if err != nil {
			return err
		}

		stmt, err := tx.PrepareContext(ctx, pq.CopyIn("blacklists", "user_id"))
		if err != nil {
			return err
		}
		defer stmt.Close()
		_, err = stmt.ExecContext(ctx, u.UserId)
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
