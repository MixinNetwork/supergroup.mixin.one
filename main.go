package main

import (
	"context"
	"flag"
	"log"

	"github.com/MixinMessenger/supergroup.mixin.one/config"
	"github.com/MixinMessenger/supergroup.mixin.one/durable"
	"github.com/MixinMessenger/supergroup.mixin.one/services"
)

func main() {
	service := flag.String("service", "http", "run a service")
	flag.Parse()

	spanner, err := durable.OpenSpannerClient(context.Background(), config.GoogleCloudSpanner)
	if err != nil {
		log.Panicln(err)
	}
	defer spanner.Close()

	switch *service {
	case "http":
		err := StartServer(spanner)
		if err != nil {
			log.Println(err)
		}
	default:
		hub := services.NewHub(spanner)
		err := hub.StartService(*service)
		if err != nil {
			log.Println(err)
		}
	}
}
