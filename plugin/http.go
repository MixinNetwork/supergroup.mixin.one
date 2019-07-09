package plugin

import (
	"fmt"
	"net/http"
)

var handlers = map[string]http.Handler{}

// called by plugin implementations
func (*PluginContext) RegisterHTTPHandler(groupName string, handler http.Handler) error {
	mutex.Lock()
	defer mutex.Unlock()

	if _, found := handlers[groupName]; found {
		return fmt.Errorf("group name already exists")
	}

	handlers[groupName] = handler
	return nil
}

// called by main codebase
func Handlers() (results map[string]http.Handler) {
	mutex.RLock()
	defer mutex.RUnlock()

	results = make(map[string]http.Handler)
	for key, value := range handlers {
		results[key] = value
	}
	return
}
