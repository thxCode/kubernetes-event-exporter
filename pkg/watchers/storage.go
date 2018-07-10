package watchers

import (
	"time"

	"k8s.io/client-go/tools/cache"
)

// WatcherStoreConfig represents the configuration of the storage backing the watcher.
type WatcherStoreConfig struct {
	KeyFunc    cache.KeyFunc
	Handler    cache.ResourceEventHandler
	StorageTTL time.Duration
}

type watcherStore struct {
	cache.Store

	handler cache.ResourceEventHandler
}

func (s *watcherStore) Add(obj interface{}) error {
	if err := s.Store.Add(obj); err != nil {
		return err
	}
	s.handler.OnAdd(obj)
	return nil
}

func (s *watcherStore) Update(obj interface{}) error {
	oldObj, _, err := s.Store.Get(obj)
	if err != nil {
		return err
	}

	if err = s.Store.Update(obj); err != nil {
		return err
	}
	s.handler.OnUpdate(oldObj, obj)
	return nil
}

func (s *watcherStore) Delete(obj interface{}) error {
	if err := s.Store.Delete(obj); err != nil {
		return err
	}
	s.handler.OnDelete(obj)
	return nil
}

func newWatcherStore(config *WatcherStoreConfig) *watcherStore {
	return &watcherStore{
		Store:   cache.NewTTLStore(config.KeyFunc, config.StorageTTL),
		handler: config.Handler,
	}
}
