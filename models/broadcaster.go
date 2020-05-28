package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup.mixin.one/session"
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

	users, err := findUsersByIdentityNumber(ctx, identity)
	if err != nil {
		return nil, err
	} else if len(users) == 0 {
		return nil, session.BadDataError(ctx)
	}
	user := users[0]

	t := time.Now()
	query := fmt.Sprintf("INSERT INTO broadcasters(user_id,created_at,updated_at) VALUES ($1,$2,$3) ON CONFLICT (user_id) DO UPDATE SET updated_at=EXCLUDED.updated_at")
	_, err = session.Database(ctx).ExecContext(ctx, query, user.UserId, t, t)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
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
