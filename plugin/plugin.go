package plugin

import (
	"log"
	"plugin"
	"sync"

	"github.com/MixinNetwork/supergroup.mixin.one/config"
	"github.com/MixinNetwork/supergroup.mixin.one/durable"
)

var (
	mutex         sync.RWMutex
	loadOnce      sync.Once
	loadedPlugins []*PluginContext
)

type PluginContext struct {
	sharedLibrary string
	config        map[string]interface{}
	plugin        *plugin.Plugin
	hostDB        *durable.Database
}

func (pc *PluginContext) load() {
	if pc.plugin == nil {
		log.Println("loading plugin", pc.sharedLibrary)
		var err error
		pc.plugin, err = plugin.Open(pc.sharedLibrary)
		if err != nil {
			log.Panicln(err)
		}

		initFunc, err := pc.plugin.Lookup("PluginInit")
		if err != nil {
			log.Panicln(err)
		}
		initFunc.(func(*PluginContext))(pc)
	}
}

func (pc *PluginContext) MixinClientID() string {
	return config.AppConfig.Mixin.ClientId
}

func (pc *PluginContext) HostDB() *durable.Database {
	return pc.hostDB
}

func (pc *PluginContext) ConfigGet(key string) (value interface{}, found bool) {
	value, found = pc.config[key]
	return
}

func (pc *PluginContext) ConfigMustGet(key string) (value interface{}) {
	return pc.config[key]
}

func LoadPlugins(database *durable.Database) {
	loadOnce.Do(func() {
		for _, p := range config.AppConfig.Plugins {
			plugCtx := &PluginContext{
				sharedLibrary: p.SharedLibrary,
				config:        p.Config,
				hostDB:        database,
			}
			plugCtx.load()
			loadedPlugins = append(loadedPlugins, plugCtx)
		}
	})
}
