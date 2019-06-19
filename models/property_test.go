package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPropertyCRUD(t *testing.T) {
	assert := assert.New(t)
	ctx := setupTestContext()
	defer teardownTestContext(ctx)

	name := "message-banned"
	p, err := CreateProperty(ctx, name, "true")
	assert.Nil(err)
	assert.NotNil(p)
	p, err = ReadProperty(ctx, name)
	assert.Nil(err)
	assert.NotNil(p)
	assert.Equal("true", p.Value)
	p, err = CreateProperty(ctx, name, "false")
	assert.Nil(err)
	assert.NotNil(p)
	p, err = ReadProperty(ctx, name)
	assert.Nil(err)
	assert.NotNil(p)
	assert.Equal("false", p.Value)
}
