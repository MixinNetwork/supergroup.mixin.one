package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client/v2"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/lib/pq"
)

type Session struct {
	UserID    string
	SessionID string
	PublicKey string
	UpdatedAt time.Time
}

var sessionsCols = []string{"user_id", "session_id", "public_key", "updated_at"}

func (s *Session) values() []interface{} {
	return []interface{}{s.UserID, s.SessionID, s.PublicKey, s.UpdatedAt}
}

func sessionFromRow(row durable.Row) (*Session, error) {
	var s Session
	err := row.Scan(&s.UserID, &s.SessionID, &s.PublicKey, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func SyncSession(ctx context.Context, sessions []*Session) error {
	if len(sessions) < 1 {
		return nil
	}
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

		repeatIds := make(map[string]bool)
		for i, s := range sessions {
			if s.SessionID == "" {
				continue
			}
			if repeatIds[s.UserID+s.SessionID] {
				continue
			}
			_, err = stmt.Exec(sessions[i].values()...)
			if err != nil {
				return err
			}
			repeatIds[s.UserID+s.SessionID] = true
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

type SimpleUser struct {
	Category string
	Sessions []*Session
}

func ReadSessionSetByUsers(ctx context.Context, userIDs []string) (map[string]*SimpleUser, error) {
	query := fmt.Sprintf("SELECT %s FROM sessions WHERE user_id=ANY($1)", strings.Join(sessionsCols, ","))

	rows, err := session.Database(ctx).QueryContext(ctx, query, pq.Array(userIDs))
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	defer rows.Close()

	set := make(map[string]*SimpleUser)
	for rows.Next() {
		s, err := sessionFromRow(rows)
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		}
		if set[s.UserID] == nil {
			su := &SimpleUser{
				Category: UserCategoryEncrypted,
				Sessions: []*Session{s},
			}
			if s.PublicKey == "" {
				su.Category = UserCategoryPlain
			}
			set[s.UserID] = su
			continue
		}
		if s.PublicKey == "" {
			set[s.UserID].Category = UserCategoryPlain
		}
		set[s.UserID].Sessions = append(set[s.UserID].Sessions, s)
	}
	return set, nil
}

func GenerateUserChecksum(sessions []*Session) string {
	if len(sessions) < 1 {
		return ""
	}
	ss := make([]*bot.Session, len(sessions))
	for i, s := range sessions {
		ss[i] = &bot.Session{
			UserID:    s.UserID,
			SessionID: s.SessionID,
			PublicKey: s.PublicKey,
		}
	}
	return bot.GenerateUserChecksum(ss)
}
