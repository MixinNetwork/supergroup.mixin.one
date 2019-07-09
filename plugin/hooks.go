package plugin

type EventType string

const (
	EventTypeMessageCreated          EventType = "MessageCreated"          // payload is github.com/MixinNetwork/supergroup.mixin.one/models.Message
	EventTypeProhibitedStatusChanged EventType = "ProhibitedStatusChanged" // payload is bool
)

var callbacks = map[EventType][]func(interface{}){}

// called by plugin implementations
func (*PluginContext) On(eventName EventType, fn func(interface{})) {
	mutex.RLock()
	defer mutex.RUnlock()

	cs, found := callbacks[eventName]
	if !found {
		cs = []func(interface{}){}
	}
	cs = append(cs, fn)
	callbacks[eventName] = cs
}

// called by main supergroup codebase
func Trigger(eventName EventType, obj interface{}) {
	mutex.RLock()
	defer mutex.RUnlock()

	cs, found := callbacks[eventName]
	if !found {
		return
	}

	for _, callback := range cs {
		callback(obj)
	}
}
