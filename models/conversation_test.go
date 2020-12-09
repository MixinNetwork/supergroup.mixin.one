package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConversationCRUD(t *testing.T) {
	assert := assert.New(t)
	ctx := setupTestContext()
	defer teardownTestContext(ctx)

	id := "0d94062e-2fef-3047-95b8-577b5b2de55a"
	err := SyncConversationParticipant(ctx, id)
	assert.Nil(err)

	sessions, err := ReadConversationParticipantSessions(ctx, id)
	assert.Nil(err)
	assert.Len(sessions, 2)

	assert.Equal("b427ca68b4787d856eb4cddb8c5473d5", generateConversationChecksum(sessions))
}
