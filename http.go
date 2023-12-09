package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MixinNetwork/bot-api-go-client/v2"
	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/middlewares"
	"github.com/MixinNetwork/supergroup.mixin.one/routes"
	"github.com/dimfeld/httptreemux"
	"github.com/gorilla/handlers"
	"github.com/unrolled/render"
)

func StartServer(database *durable.Database) error {
	mixin := config.AppConfig.Mixin
	_, err := bot.UpdatePreference(context.Background(), mixin.ClientId, mixin.SessionId, mixin.SessionKey, "", "CONTACTS", "", 0)
	if err != nil {
		return err
	}

	logger := durable.NewLoggerClient()
	router := httptreemux.New()
	routes.RegisterRoutes(router)
	handler := middlewares.Authenticate(router)
	handler = middlewares.Constraint(handler)
	handler = middlewares.Context(handler, database, render.New())
	handler = middlewares.Stats(handler, "http", true, config.BuildVersion)
	handler = middlewares.Log(handler, logger, "http")
	handler = handlers.ProxyHeaders(handler)

	return http.ListenAndServe(fmt.Sprintf(":%d", config.AppConfig.Service.HTTPListenPort), handler)
}
