package plugin

import (
	"log"
	"plugin"
	"sync"

	"github.com/MixinNetwork/supergroup.mixin.one/config"
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

func (pc *PluginContext) ConfigGet(key string) (value interface{}, found bool) {
	value, found = pc.config[key]
	return
}

func (pc *PluginContext) ConfigMustGet(key string) (value interface{}) {
	return pc.config[key]
}

func LoadPlugins() {
	loadOnce.Do(func() {
		for _, p := range config.AppConfig.Plugins {
			plugCtx := &PluginContext{
				sharedLibrary: p.SharedLibrary,
				config:        p.Config,
			}
			plugCtx.load()
			loadedPlugins = append(loadedPlugins, plugCtx)
		}
	})
}
