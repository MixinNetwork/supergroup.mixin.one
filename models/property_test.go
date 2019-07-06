package models

import (
	"context"
	"database/sql"
	"testing"

	"github.com/MixinNetwork/supergroup.mixin.one/session"
	"github.com/stretchr/testify/assert"
)

func TestPropertyCRUD(t *testing.T) {
	assert := assert.New(t)
	ctx := setupTestContext()
	defer teardownTestContext(ctx)

	name := ProhibitedMessage
	b, err := testReadPropertyAsBool(ctx, name)
	assert.False(b)
	assert.Nil(err)
	p, err := CreateProperty(ctx, name, true)
	assert.Nil(err)
	assert.NotNil(p)
	p, err = ReadProperty(ctx, name)
	assert.Nil(err)
	assert.NotNil(p)
	assert.Equal("true", p.Value)
	b, err = ReadProhibitedProperty(ctx)
	assert.Nil(err)
	assert.True(b)
	b, err = testReadPropertyAsBool(ctx, name)
	assert.True(b)
	assert.Nil(err)
	p, err = CreateProperty(ctx, name, false)
	assert.Nil(err)
	assert.NotNil(p)
	p, err = ReadProperty(ctx, name)
	assert.Nil(err)
	assert.NotNil(p)
	assert.Equal("false", p.Value)
	b, err = testReadPropertyAsBool(ctx, name)
	assert.False(b)
	assert.Nil(err)
	b, err = ReadProhibitedProperty(ctx)
	assert.Nil(err)
	assert.False(b)
}

func testReadPropertyAsBool(ctx context.Context, name string) (bool, error) {
	var b bool
	err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		b, err = readPropertyAsBool(ctx, tx, name)
		return err
	})
	return b, err
}
