package plugin

import "sync"

var (
	mutex sync.RWMutex
)

const (
	PluginsDir = "./plugins"
)
