package events

import (
	"github.com/sirupsen/logrus"
	apiCoreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

// EventHandler interface provides a way to act upon signals
// from watcher that only watches the events resource.
type EventHandler interface {
	OnAdd(event *apiCoreV1.Event)
	OnUpdate(oldEvent *apiCoreV1.Event, newEvent *apiCoreV1.Event)
	OnDelete(*apiCoreV1.Event)
}

type eventHandlerWrapper struct {
	handler EventHandler
}

func NewEventHandlerWrapper(handler EventHandler) *eventHandlerWrapper {
	return &eventHandlerWrapper{
		handler: handler,
	}
}

func (c *eventHandlerWrapper) OnAdd(obj interface{}) {
	if event, ok := c.convert(obj); ok {
		c.handler.OnAdd(event)
	}
}

func (c *eventHandlerWrapper) OnUpdate(oldObj interface{}, newObj interface{}) {
	oldEvent, oldOk := c.convert(oldObj)
	newEvent, newOk := c.convert(newObj)
	if newOk && (oldObj == nil || oldOk) {
		c.handler.OnUpdate(oldEvent, newEvent)
	}
}

func (c *eventHandlerWrapper) OnDelete(obj interface{}) {
	event, ok := obj.(*apiCoreV1.Event)

	// When a delete is dropped, the relist will notice a pod in the store not
	// in the list, leading to the insertion of a tombstone object which contains
	// the deleted key/value. Note that this value might be stale.
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			logrus.Warnf("Object is neither event nor tombstone: %+v", obj)
			return
		}
		event, ok = tombstone.Obj.(*apiCoreV1.Event)
		if !ok {
			logrus.Warnf("Tombstone contains object that is not a event: %+v", obj)
			return
		}
	}

	c.handler.OnDelete(event)
}

func (c *eventHandlerWrapper) convert(obj interface{}) (*apiCoreV1.Event, bool) {
	if event, ok := obj.(*apiCoreV1.Event); ok {
		return event, true
	}

	logrus.Warnf("Event watch handler received not event, but %+v", obj)
	return nil, false
}
