package v3

import (
	"context"

	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var (
	HuaWeiClusterEventGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "HuaWeiClusterEvent",
	}
	HuaWeiClusterEventResource = metav1.APIResource{
		Name:         "huaweiclusterevents",
		SingularName: "huaweiclusterevent",
		Namespaced:   false,
		Kind:         HuaWeiClusterEventGroupVersionKind.Kind,
	}
)

type HuaWeiClusterEventList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HuaWeiClusterEvent
}

type HuaWeiClusterEventHandlerFunc func(key string, obj *HuaWeiClusterEvent) error

type HuaWeiClusterEventLister interface {
	List(namespace string, selector labels.Selector) (ret []*HuaWeiClusterEvent, err error)
	Get(namespace, name string) (*HuaWeiClusterEvent, error)
}

type HuaWeiClusterEventController interface {
	Informer() cache.SharedIndexInformer
	Lister() HuaWeiClusterEventLister
	AddHandler(name string, handler HuaWeiClusterEventHandlerFunc)
	AddClusterScopedHandler(name, clusterName string, handler HuaWeiClusterEventHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type HuaWeiClusterEventInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*HuaWeiClusterEvent) (*HuaWeiClusterEvent, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*HuaWeiClusterEvent, error)
	Get(name string, opts metav1.GetOptions) (*HuaWeiClusterEvent, error)
	Update(*HuaWeiClusterEvent) (*HuaWeiClusterEvent, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*HuaWeiClusterEventList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() HuaWeiClusterEventController
	AddHandler(name string, sync HuaWeiClusterEventHandlerFunc)
	AddLifecycle(name string, lifecycle HuaWeiClusterEventLifecycle)
	AddClusterScopedHandler(name, clusterName string, sync HuaWeiClusterEventHandlerFunc)
	AddClusterScopedLifecycle(name, clusterName string, lifecycle HuaWeiClusterEventLifecycle)
}

type huaWeiClusterEventLister struct {
	controller *huaWeiClusterEventController
}

func (l *huaWeiClusterEventLister) List(namespace string, selector labels.Selector) (ret []*HuaWeiClusterEvent, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*HuaWeiClusterEvent))
	})
	return
}

func (l *huaWeiClusterEventLister) Get(namespace, name string) (*HuaWeiClusterEvent, error) {
	var key string
	if namespace != "" {
		key = namespace + "/" + name
	} else {
		key = name
	}
	obj, exists, err := l.controller.Informer().GetIndexer().GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(schema.GroupResource{
			Group:    HuaWeiClusterEventGroupVersionKind.Group,
			Resource: "huaWeiClusterEvent",
		}, key)
	}
	return obj.(*HuaWeiClusterEvent), nil
}

type huaWeiClusterEventController struct {
	controller.GenericController
}

func (c *huaWeiClusterEventController) Lister() HuaWeiClusterEventLister {
	return &huaWeiClusterEventLister{
		controller: c,
	}
}

func (c *huaWeiClusterEventController) AddHandler(name string, handler HuaWeiClusterEventHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*HuaWeiClusterEvent))
	})
}

func (c *huaWeiClusterEventController) AddClusterScopedHandler(name, cluster string, handler HuaWeiClusterEventHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}

		if !controller.ObjectInCluster(cluster, obj) {
			return nil
		}

		return handler(key, obj.(*HuaWeiClusterEvent))
	})
}

type huaWeiClusterEventFactory struct {
}

func (c huaWeiClusterEventFactory) Object() runtime.Object {
	return &HuaWeiClusterEvent{}
}

func (c huaWeiClusterEventFactory) List() runtime.Object {
	return &HuaWeiClusterEventList{}
}

func (s *huaWeiClusterEventClient) Controller() HuaWeiClusterEventController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.huaWeiClusterEventControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(HuaWeiClusterEventGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &huaWeiClusterEventController{
		GenericController: genericController,
	}

	s.client.huaWeiClusterEventControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type huaWeiClusterEventClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   HuaWeiClusterEventController
}

func (s *huaWeiClusterEventClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *huaWeiClusterEventClient) Create(o *HuaWeiClusterEvent) (*HuaWeiClusterEvent, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*HuaWeiClusterEvent), err
}

func (s *huaWeiClusterEventClient) Get(name string, opts metav1.GetOptions) (*HuaWeiClusterEvent, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*HuaWeiClusterEvent), err
}

func (s *huaWeiClusterEventClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*HuaWeiClusterEvent, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*HuaWeiClusterEvent), err
}

func (s *huaWeiClusterEventClient) Update(o *HuaWeiClusterEvent) (*HuaWeiClusterEvent, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*HuaWeiClusterEvent), err
}

func (s *huaWeiClusterEventClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *huaWeiClusterEventClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *huaWeiClusterEventClient) List(opts metav1.ListOptions) (*HuaWeiClusterEventList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*HuaWeiClusterEventList), err
}

func (s *huaWeiClusterEventClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *huaWeiClusterEventClient) Patch(o *HuaWeiClusterEvent, data []byte, subresources ...string) (*HuaWeiClusterEvent, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*HuaWeiClusterEvent), err
}

func (s *huaWeiClusterEventClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *huaWeiClusterEventClient) AddHandler(name string, sync HuaWeiClusterEventHandlerFunc) {
	s.Controller().AddHandler(name, sync)
}

func (s *huaWeiClusterEventClient) AddLifecycle(name string, lifecycle HuaWeiClusterEventLifecycle) {
	sync := NewHuaWeiClusterEventLifecycleAdapter(name, false, s, lifecycle)
	s.AddHandler(name, sync)
}

func (s *huaWeiClusterEventClient) AddClusterScopedHandler(name, clusterName string, sync HuaWeiClusterEventHandlerFunc) {
	s.Controller().AddClusterScopedHandler(name, clusterName, sync)
}

func (s *huaWeiClusterEventClient) AddClusterScopedLifecycle(name, clusterName string, lifecycle HuaWeiClusterEventLifecycle) {
	sync := NewHuaWeiClusterEventLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.AddClusterScopedHandler(name, clusterName, sync)
}
