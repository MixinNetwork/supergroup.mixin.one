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
	return &s, err
}

func SyncSession(ctx context.Context, sessions []*Session) error {
	var userIDs, toRemove []string
	var toAdd []*Session
	toLatest := make(map[string]*Session)
	for i, s := range sessions {
		userIDs = append(userIDs, s.UserID)
		toLatest[s.SessionID] = sessions[i]
	}
	olds, err := ReadSessionsByUsers(ctx, userIDs)
	if err != nil {
		return err
	}
	existing := make(map[string]*Session)
	for i, s := range olds {
		if toLatest[s.SessionID] == nil {
			toRemove = append(toRemove, s.SessionID)
		}
		existing[s.SessionID] = olds[i]
	}
	for i, s := range sessions {
		if existing[s.SessionID] == nil {
			toAdd = append(toAdd, sessions[i])
		}
	}

	if len(toAdd) == 0 && len(toRemove) == 0 {
		return nil
	}

	err = session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, pq.CopyIn("sessions", sessionsCols...))
		if err != nil {
			return err
		}
		defer stmt.Close()

		for i, _ := range toAdd {
			_, err = stmt.Exec(toAdd[i].values()...)
			if err != nil {
				return err
			}
		}
		_, err = stmt.Exec()
		if err != nil {
			return err
		}
		_, err = tx.Exec("DELETE FROM sessions WHERE session_id=ANY($1)", pq.Array(toRemove))
		return nil
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
