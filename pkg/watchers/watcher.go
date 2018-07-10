package watchers

import (
	"time"

	"k8s.io/client-go/tools/cache"
)

// WatcherConfig represents the configuration of the Kubernetes API watcher.
type WatcherConfig struct {
	ListerWatcher cache.ListerWatcher
	ExpectedType  interface{}
	StoreConfig   *WatcherStoreConfig
	ResyncPeriod  time.Duration
}

// Watcher is an interface of the generic proactive API watcher.
type Watcher interface {
	Run(stopCh <-chan struct{})
}

type watcher struct {
	reflector *cache.Reflector
}

func (w *watcher) Run(stopCh <-chan struct{}) {
	w.reflector.Run(stopCh)
}

// NewWatcher creates a new Kubernetes API watcher using provided configuration.
func NewWatcher(config *WatcherConfig) Watcher {
	return &watcher{
		reflector: cache.NewReflector(
			config.ListerWatcher,
			config.ExpectedType,
			newWatcherStore(config.StoreConfig),
			config.ResyncPeriod,
		),
	}
}
