package v3

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type HuaWeiEventLogLifecycle interface {
	Create(obj *HuaWeiEventLog) (*HuaWeiEventLog, error)
	Remove(obj *HuaWeiEventLog) (*HuaWeiEventLog, error)
	Updated(obj *HuaWeiEventLog) (*HuaWeiEventLog, error)
}

type huaWeiEventLogLifecycleAdapter struct {
	lifecycle HuaWeiEventLogLifecycle
}

func (w *huaWeiEventLogLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*HuaWeiEventLog))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *huaWeiEventLogLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*HuaWeiEventLog))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *huaWeiEventLogLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*HuaWeiEventLog))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewHuaWeiEventLogLifecycleAdapter(name string, clusterScoped bool, client HuaWeiEventLogInterface, l HuaWeiEventLogLifecycle) HuaWeiEventLogHandlerFunc {
	adapter := &huaWeiEventLogLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *HuaWeiEventLog) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
