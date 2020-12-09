package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/lib/pq"
)

type Session struct {
	UserID    string
	SessionID string
	PublicKey string
}

var sessionsCols = []string{"user_id", "session_id", "public_key"}

func (s *Session) values() []interface{} {
	return []interface{}{s.UserID, s.SessionID, s.PublicKey}
}

func sessionFromRow(row durable.Row) (*Session, error) {
	var s Session
	err := row.Scan(&s.UserID, &s.SessionID, &s.PublicKey)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func SyncSession(ctx context.Context, sessions []*Session) error {
	var userIDs []string
	for _, s := range sessions {
		userIDs = append(userIDs, s.UserID)
	}

	err := session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.Exec("DELETE FROM sessions WHERE user_id=ANY($1)", pq.Array(userIDs))
		if err != nil {
			return err
		}
		stmt, err := tx.PrepareContext(ctx, pq.CopyIn("sessions", sessionsCols...))
		if err != nil {
			return err
		}
		defer stmt.Close()

		for i, _ := range sessions {
			_, err = stmt.Exec(sessions[i].values()...)
			if err != nil {
				return err
			}
		}
		_, err = stmt.Exec()
		return err
	})
	if err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

func ReadSessionsByUsers(ctx context.Context, userIDs []string) ([]*Session, error) {
	var sessions []*Session
	query := fmt.Sprintf("SELECT %s FROM sessions WHERE user_id=ANY($1)", strings.Join(sessionsCols, ","))

	rows, err := session.Database(ctx).QueryContext(ctx, query, pq.Array(userIDs))
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	defer rows.Close()

	for rows.Next() {
		s, err := sessionFromRow(rows)
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}
