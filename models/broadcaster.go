package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/lib/pq"
)

type Broadcaster struct {
	UserId    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

var broadcasterColumns = []string{"user_id", "created_at", "updated_at"}

func (b *Broadcaster) values() []interface{} {
	return []interface{}{b.UserId, b.CreatedAt, b.UpdatedAt}
}

func (current *User) CreateBroadcaster(ctx context.Context, identity int64) (*User, error) {
	if !current.isAdmin() {
		return nil, session.ForbiddenError(ctx)
	}

	user, err := findUserByIdentityNumber(ctx, identity)
	if err != nil || user == nil {
		return nil, err
	}

	t := time.Now()
	b := &Broadcaster{
		UserId:    user.UserId,
		CreatedAt: t,
		UpdatedAt: t,
	}

	err = session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, pq.CopyIn("broadcasters", broadcasterColumns...))
		if err != nil {
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(b.values()...)
		return err
	})
	if err != nil {
		return user, session.TransactionError(ctx, err)
	}
	return user, nil
}

func ReadBroadcasters(ctx context.Context) ([]*User, error) {
	query := fmt.Sprintf("SELECT %s FROM users WHERE user_id IN (SELECT user_id FROM broadcasters ORDER BY updated_at DESC LIMIT 5)", strings.Join(usersCols, ","))
	rows, err := session.Database(ctx).QueryContext(ctx, query)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u, err := userFromRow(rows)
		if err != nil {
			return users, session.TransactionError(ctx, err)
		}
		users = append(users, u)
	}
	return users, nil
}
