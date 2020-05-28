package models

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

const (
	testEnvironment = "test"
	testDatabase    = "group_test"
)

const (
	dropRewardsDDL             = `DROP TABLE IF EXISTS rewards;`
	dropBroadcastersDDL        = `DROP TABLE IF EXISTS broadcasters;`
	dropPropertiesDDL          = `DROP TABLE IF EXISTS properties;`
	dropParticipantsDDL        = `DROP TABLE IF EXISTS participants;`
	dropPacketsDDL             = `DROP TABLE IF EXISTS packets;`
	dropAssetsDDL              = `DROP TABLE IF EXISTS assets;`
	dropBlacklistsDDL          = `DROP TABLE IF EXISTS blacklists;`
	dropDistributedMessagesDDL = `DROP TABLE IF EXISTS distributed_messages;`
	dropMessagesDDL            = `DROP TABLE IF EXISTS messages;`
	dropUsersDDL               = `DROP TABLE IF EXISTS users;`
)

func TestClear(t *testing.T) {
	ctx := setupTestContext()
	teardownTestContext(ctx)
}

func teardownTestContext(ctx context.Context) {
	db := session.Database(ctx)
	tables := []string{
		dropUsersDDL,
		dropMessagesDDL,
		dropDistributedMessagesDDL,
		dropBlacklistsDDL,
		dropAssetsDDL,
		dropParticipantsDDL,
		dropPacketsDDL,
		dropPropertiesDDL,
		dropBroadcastersDDL,
		dropRewardsDDL,
	}
	for _, q := range tables {
		if _, err := db.Exec(q); err != nil {
			log.Panicln(err)
		}
	}
}

func setupTestContext() context.Context {
	config.LoadConfig("../config")
	if config.AppConfig.Service.Environment != testEnvironment || config.AppConfig.Database.Name != testDatabase {
		log.Panicln(config.AppConfig.Service.Environment, config.AppConfig.Database.Name)
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", config.AppConfig.Database.User, config.AppConfig.Database.Password, config.AppConfig.Database.Host, config.AppConfig.Database.Port, config.AppConfig.Database.Name)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Panicln(err)
	}
	data, err := ioutil.ReadFile("./schema.sql")
	if _, err := db.Exec(string(data)); err != nil {
		log.Panicln(err)
	}
	database, err := durable.NewDatabase(context.Background(), db)
	if err != nil {
		log.Panicln(err)
	}
	return session.WithDatabase(context.Background(), database)
}
