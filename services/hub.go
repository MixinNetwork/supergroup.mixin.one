package services

import (
	"context"
	"fmt"

	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/session"
)

type Hub struct {
	context  context.Context
	services map[string]Service
}

func NewHub(db *durable.Database) *Hub {
	hub := &Hub{services: make(map[string]Service)}
	hub.context = session.WithDatabase(context.Background(), db)
	hub.registerServices()
	return hub
}

func (hub *Hub) StartService(name string) error {
	service := hub.services[name]
	if service == nil {
		return fmt.Errorf("no service found: %s", name)
	}

	ctx := session.WithLogger(hub.context, durable.BuildLogger())
	return service.Run(ctx)
}

func (hub *Hub) registerServices() {
	hub.services["message"] = &MessageService{}
}
