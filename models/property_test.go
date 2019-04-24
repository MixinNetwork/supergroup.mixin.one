package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPropertyCRUD(t *testing.T) {
	assert := assert.New(t)
	ctx := setupTestContext()
	defer teardownTestContext(ctx)

	key := "asset"
	tm := time.Now()
	err := WriteProperty(ctx, key, tm.Format(time.RFC3339Nano))
	assert.Nil(err)
	str, err := ReadProperty(ctx, key)
	assert.Nil(err)
	assert.Equal(tm.Format(time.RFC3339Nano), str)
	date, err := ReadPropertyAsOffset(ctx, key)
	assert.Nil(err)
	assert.True(tm.Equal(date))
	tm2 := time.Now()
	err = WriteProperty(ctx, key, tm2.Format(time.RFC3339Nano))
	assert.Nil(err)
	date, err = ReadPropertyAsOffset(ctx, key)
	assert.Nil(err)
	assert.True(tm2.Equal(date))
}
