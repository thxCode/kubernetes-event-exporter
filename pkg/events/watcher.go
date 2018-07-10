package events

import (
	"time"

	apiCoreV1 "k8s.io/api/core/v1"
	apisMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/thxcode/kubernetes-event-exporter/pkg/watchers"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// OnListFunc represent an action on the initial list of object received
// from the Kubernetes API server before starting watching for the updates.
type OnListFunc func(*apiCoreV1.EventList)

// EventWatcherConfig represents the configuration for the watcher that
// only watches the events resource.
type EventWatcherConfig struct {
	// Note, that this action will be executed on each List request, of which
	// there can be many, e.g. because of network problems. Note also, that
	// items in the List response WILL NOT trigger OnAdd method in handler,
	// instead Store contents will be completely replaced.
	OnList       OnListFunc
	ResyncPeriod time.Duration
	StorageTTL   time.Duration
	Handler      EventHandler
}

// NewEventWatcher create a new watcher that only watches the events resource.
func NewEventWatcher(client kubernetes.Interface, config *EventWatcherConfig) watchers.Watcher {
	return watchers.NewWatcher(&watchers.WatcherConfig{
		ListerWatcher: &cache.ListWatch{
			DisableChunking: true,
			ListFunc: func(options apisMetaV1.ListOptions) (runtime.Object, error) {
				list, err := client.CoreV1().Events(apisMetaV1.NamespaceAll).List(options)
				if err == nil {
					config.OnList(list)
				}
				return list, err
			},
			WatchFunc: func(options apisMetaV1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Events(apisMetaV1.NamespaceAll).Watch(options)
			},
		},
		ExpectedType: &apiCoreV1.Event{},
		StoreConfig: &watchers.WatcherStoreConfig{
			KeyFunc:    cache.DeletionHandlingMetaNamespaceKeyFunc,
			Handler:    NewEventHandlerWrapper(config.Handler),
			StorageTTL: config.StorageTTL,
		},
		ResyncPeriod: config.ResyncPeriod,
	})
}
