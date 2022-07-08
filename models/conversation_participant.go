package models

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/externals"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/lib/pq"
)

type ConversationParticipant struct {
	ConversationID string
	UserID         string
}

var conversationParticipantsCols = []string{"conversation_id", "user_id"}

func (p *ConversationParticipant) values() []interface{} {
	return []interface{}{p.ConversationID, p.UserID}
}

func conversationParticipantFromRow(row durable.Row) (*ConversationParticipant, error) {
	var p ConversationParticipant
	err := row.Scan(&p.ConversationID, &p.UserID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func SyncConversationParticipant(ctx context.Context, conversationID string) error {
	conversation, err := externals.ReadConversation(ctx, conversationID)
	if err != nil || conversation == nil {
		return err
	}
	var sessions []*Session
	for _, ps := range conversation.ParticipantSessions {
		sessions = append(sessions, &Session{
			UserID:    ps.UserId,
			SessionID: ps.SessionId,
			PublicKey: ps.PublicKey,
			UpdatedAt: time.Now(),
		})
	}

	err = session.Database(ctx).RunInTransaction(ctx, nil, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.Exec("DELETE FROM conversation_participants WHERE conversation_id=$1", conversationID)
		if err != nil {
			return err
		}
		stmt, err := tx.PrepareContext(ctx, pq.CopyIn("conversation_participants", conversationParticipantsCols...))
		if err != nil {
			return err
		}
		defer stmt.Close()
		for _, p := range conversation.Participants {
			_, err = stmt.Exec(conversationID, p.UserId)
			if err != nil {
				return err
			}
		}
		_, err = stmt.Exec()
		return nil
	})
	if err != nil {
		return session.TransactionError(ctx, err)
	}
	return SyncSession(ctx, sessions)
}

func ReadConversationParticipantSessions(ctx context.Context, conversationID string) ([]*Session, error) {
	query := fmt.Sprintf("SELECT %s FROM sessions WHERE user_id IN (SELECT user_id FROM conversation_participants WHERE conversation_id=$1)", strings.Join(sessionsCols, ","))

	rows, err := session.Database(ctx).QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		s, err := sessionFromRow(rows)
		if err != nil {
			return nil, session.TransactionError(ctx, err)
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func generateConversationChecksum(devices []*Session) string {
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].SessionID < devices[j].SessionID
	})
	h := md5.New()
	for _, d := range devices {
		io.WriteString(h, d.SessionID)
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum[:])
}
