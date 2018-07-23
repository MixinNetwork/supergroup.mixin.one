package models

import (
	"context"
	"log"
	"testing"

	"cloud.google.com/go/spanner"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

const (
	testEnvironment = "test"
	testDatabase    = "projects/mixin-183904/instances/group-assistant/databases/test"
)

func TestClear(t *testing.T) {
	ctx := setupTestContext()
	teardownTestContext(ctx)
}

func teardownTestContext(ctx context.Context) {
	db := session.Database(ctx)
	tables := []string{
		"users",
		"messages",
		"distributed_messages",
		"properties",
	}
	for _, table := range tables {
		db.Apply(ctx, []*spanner.Mutation{spanner.Delete(table, spanner.AllKeys())}, "all", "DELETE", "DELETE FROM all")
	}
	db.Close()
}

func setupTestContext() context.Context {
	if config.Environment != testEnvironment || config.GoogleCloudSpanner != testDatabase {
		log.Panicln(config.Environment, config.GoogleCloudSpanner)
	}

	spanner, err := durable.OpenSpannerClient(context.Background(), config.GoogleCloudSpanner)
	if err != nil {
		log.Panicln(err)
	}

	db := durable.WrapDatabase(spanner, nil)
	return session.WithDatabase(context.Background(), db)
}
