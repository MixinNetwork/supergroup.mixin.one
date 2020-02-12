package main

import (
	"fmt"
	"net/http"

	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/middlewares"
	"github.com/MixinNetwork/supergroup.mixin.one/routes"
	"github.com/dimfeld/httptreemux"
	"github.com/gorilla/handlers"
	"github.com/unrolled/render"
)

func StartServer(database *durable.Database) error {
	logger := durable.NewLoggerClient()
	router := httptreemux.New()
	routes.RegisterRoutes(router)
	handler := middlewares.Authenticate(router)
	handler = middlewares.Constraint(handler)
	handler = middlewares.Context(handler, database, render.New(render.Options{UnEscapeHTML: true}))
	handler = middlewares.Stats(handler, "http", true, config.BuildVersion)
	handler = middlewares.Log(handler, logger, "http")
	handler = handlers.ProxyHeaders(handler)

	return http.ListenAndServe(fmt.Sprintf(":%d", config.AppConfig.Service.HTTPListenPort), handler)
}
