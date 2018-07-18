package models

import (
	"context"
	"crypto/md5"
	"io"
	"strings"

	"cloud.google.com/go/spanner"
	"github.com/MixinMessenger/supergroup.mixin.one/durable"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/api/iterator"
)

func UniqueConversationId(userId, recipientId string) string {
	minId, maxId := userId, recipientId
	if strings.Compare(userId, recipientId) > 0 {
		maxId, minId = userId, recipientId
	}
	h := md5.New()
	io.WriteString(h, minId)
	io.WriteString(h, maxId)
	sum := h.Sum(nil)
	sum[6] = (sum[6] & 0x0f) | 0x30
	sum[8] = (sum[8] & 0x3f) | 0x80
	return uuid.FromBytesOrNil(sum).String()
}

func readCollectionIds(ctx context.Context, txn durable.Transaction, query string, params map[string]interface{}) ([]string, error) {
	it := txn.Query(ctx, spanner.Statement{SQL: query, Params: params})
	defer it.Stop()

	var ids []string
	for {
		row, err := it.Next()
		if err == iterator.Done {
			return ids, nil
		} else if err != nil {
			return ids, err
		}
		var id string
		err = row.Columns(&id)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}
}
