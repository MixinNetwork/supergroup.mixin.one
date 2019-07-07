package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"plugin"
	"regexp"
	"time"

	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
	"github.com/MixinNetwork/supergroup.mixin.one/services"
)

func main() {
	service := flag.String("service", "http", "run a service")
	dir := flag.String("dir", "./", "config.yaml dir")
	flag.Parse()

	loadPlugins()

	config.LoadConfig(*dir)
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.AppConfig.Database.DatebaseUser,
		config.AppConfig.Database.DatabasePassword,
		config.AppConfig.Database.DatabaseHost,
		config.AppConfig.Database.DatabasePort,
		config.AppConfig.Database.DatabaseName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Panicln(err)
	}
	db.SetConnMaxLifetime(time.Hour)
	db.SetMaxOpenConns(128)
	db.SetMaxIdleConns(4)

	defer db.Close()
	database, err := durable.NewDatabase(context.Background(), db)
	if err != nil {
		log.Panicln(err)
	}

	switch *service {
	case "http":
		if config.AppConfig.System.AccpetWeChatPayment {
			go services.StartWxPaymentWatch(*service, database)
		}
		err := StartServer(database)
		if err != nil {
			log.Println(err)
		}
	default:
		go func() {
			hub := services.NewHub(database)
			err := hub.StartService(*service)
			if err != nil {
				log.Println(err)
			}
		}()
		http.ListenAndServe(fmt.Sprintf(":%d", config.AppConfig.Service.HTTPListenPort+2000), http.DefaultServeMux)
	}
}

func loadPlugins() {
	pluginsDir := "./plugins"
	if _, err := os.Stat(pluginsDir); err != nil {
		return
	}
	files, err := ioutil.ReadDir(pluginsDir)
	if err != nil {
		log.Panicln(err)
	}

	plugins := []os.FileInfo{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		matched, err := regexp.MatchString(".*\\.so", file.Name())
		if err != nil {
			log.Panicln(err)
		}
		if matched {
			plugins = append(plugins, file)
		}
	}

	for _, pluginFile := range plugins {
		log.Println("loading plugin", pluginFile.Name())
		_, err := plugin.Open(path.Join(pluginsDir, pluginFile.Name()))
		if err != nil {
			log.Panicln(err)
		}
	}
}
