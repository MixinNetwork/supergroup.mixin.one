package services

import (
	"context"
	"log"

	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

type ServiceAll struct{}

func NewServiceAll() *ServiceAll {
	return &ServiceAll{}
}

func (service *ServiceAll) Run(db *durable.Database) error {
	log.Println("running all service")
	ctx := session.WithDatabase(context.Background(), db)
	ctx = session.WithLogger(ctx, durable.BuildLogger())
	go distribute(ctx)
	go loopInactiveUsers(ctx)
	go loopPendingMessages(ctx)
	go handlePendingParticipants(ctx)
	go handleExpiredPackets(ctx)
	go handlePendingRewards(ctx)
	loopPendingSuccessMessages(ctx)
	return nil
}
