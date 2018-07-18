package services

import (
	"fmt"
	"time"

	"github.com/MixinMessenger/supergroup.mixin.one/durable"
)

type Event struct {
	id      string
	time    time.Time
	cost    time.Duration
	err     error
	factory *EventFactory
}

type EventFactory struct {
	pool       chan Event
	fa         chan Event
	ga         chan Event
	size       int64
	processed  int64
	processing int64
	failed     int64
	time       time.Time
	cost       time.Duration
}

func NewEventFactory(size int64) *EventFactory {
	factory := &EventFactory{
		pool:       make(chan Event, size),
		fa:         make(chan Event, size),
		ga:         make(chan Event, size),
		time:       time.Now(),
		size:       size,
		processed:  0,
		processing: 0,
		cost:       0,
	}
	for i := int64(0); i < size; i++ {
		factory.pool <- Event{factory: factory}
	}
	go func() {
		for {
			select {
			case event := <-factory.fa:
				if event.err != nil {
					factory.failed = factory.failed + 1
				}
				factory.processing = factory.processing - 1
				factory.processed = factory.processed + 1
				factory.cost = factory.cost + event.cost
			case <-factory.ga:
				factory.processing = factory.processing + 1
			}
		}
	}()
	return factory
}

func (factory *EventFactory) InsightRoutine(logger *durable.Logger, tag string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Infof("%s %s", tag, factory.Insight())
		}
	}
}

func (factory *EventFactory) Insight() string {
	running := time.Now().Sub(factory.time).Minutes()
	average := factory.cost.Seconds() / float64(factory.processed)
	return fmt.Sprintf("RUNNING FOR %f MINUTES, PROCESSED %d/%d EVENTS, TOTAL COSTS %f MINUTES, AVERAGE COST %f SECONDS, PROCESSING %d/%d, CAPACITY %f TPS",
		running, factory.failed, factory.processed, factory.cost.Minutes(), average, factory.processing, factory.size, float64(factory.processed)/running/60)
}

func (factory *EventFactory) Get() Event {
	event := <-factory.pool
	event.time = time.Now()
	factory.ga <- event
	return event
}

func (event Event) Finalize(err error) error {
	event.err = err
	event.cost = time.Now().Sub(event.time)
	event.factory.pool <- event
	event.factory.fa <- event
	return err
}
