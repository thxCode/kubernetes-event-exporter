package v3

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type HuaWeiClusterEventLifecycle interface {
	Create(obj *HuaWeiClusterEvent) (*HuaWeiClusterEvent, error)
	Remove(obj *HuaWeiClusterEvent) (*HuaWeiClusterEvent, error)
	Updated(obj *HuaWeiClusterEvent) (*HuaWeiClusterEvent, error)
}

type huaWeiClusterEventLifecycleAdapter struct {
	lifecycle HuaWeiClusterEventLifecycle
}

func (w *huaWeiClusterEventLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*HuaWeiClusterEvent))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *huaWeiClusterEventLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*HuaWeiClusterEvent))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *huaWeiClusterEventLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*HuaWeiClusterEvent))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewHuaWeiClusterEventLifecycleAdapter(name string, clusterScoped bool, client HuaWeiClusterEventInterface, l HuaWeiClusterEventLifecycle) HuaWeiClusterEventHandlerFunc {
	adapter := &huaWeiClusterEventLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *HuaWeiClusterEvent) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
