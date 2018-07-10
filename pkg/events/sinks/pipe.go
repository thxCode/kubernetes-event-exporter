package sinks

import (
	apiCoreV1 "k8s.io/api/core/v1"
)

type Pipe interface {
	Start() error
	Stop()

	OnAdd(event *apiCoreV1.Event) error
	OnUpdate(oldEvent *apiCoreV1.Event, newEvent *apiCoreV1.Event) error
	OnDelete(event *apiCoreV1.Event) error
	OnList(eventList *apiCoreV1.EventList) error
}
