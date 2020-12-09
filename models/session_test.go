package models

import (
	"testing"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/stretchr/testify/assert"
)

func TestSessionCRUD(t *testing.T) {
	assert := assert.New(t)
	ctx := setupTestContext()
	defer teardownTestContext(ctx)

	uid := bot.UuidNewV4().String()
	sessions := []*Session{
		&Session{
			UserID:    uid,
			SessionID: bot.UuidNewV4().String(),
		},
	}

	err := SyncSession(ctx, sessions)
	assert.Nil(err)
	sessions, err = ReadSessionsByUsers(ctx, []string{uid})
	assert.Nil(err)
	assert.Len(sessions, 1)
	err = SyncSession(ctx, sessions)
	assert.Nil(err)
	sessions, err = ReadSessionsByUsers(ctx, []string{uid})
	assert.Nil(err)
	assert.Len(sessions, 1)

	uid2 := bot.UuidNewV4().String()
	sessions = []*Session{
		&Session{
			UserID:    uid2,
			SessionID: bot.UuidNewV4().String(),
		},
		&Session{
			UserID:    uid,
			SessionID: bot.UuidNewV4().String(),
		},
		&Session{
			UserID:    uid,
			SessionID: bot.UuidNewV4().String(),
		},
	}
	assert.NotEqual(uid2, uid)
	err = SyncSession(ctx, sessions)
	assert.Nil(err)
	sessions, err = ReadSessionsByUsers(ctx, []string{uid2})
	assert.Nil(err)
	assert.Len(sessions, 1)
	sessions, err = ReadSessionsByUsers(ctx, []string{uid})
	assert.Nil(err)
	assert.Len(sessions, 2)
}
