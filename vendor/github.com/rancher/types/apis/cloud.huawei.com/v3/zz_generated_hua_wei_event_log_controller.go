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
	HuaWeiEventLogGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "HuaWeiEventLog",
	}
	HuaWeiEventLogResource = metav1.APIResource{
		Name:         "huaweieventlogs",
		SingularName: "huaweieventlog",
		Namespaced:   true,

		Kind: HuaWeiEventLogGroupVersionKind.Kind,
	}
)

type HuaWeiEventLogList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HuaWeiEventLog
}

type HuaWeiEventLogHandlerFunc func(key string, obj *HuaWeiEventLog) error

type HuaWeiEventLogLister interface {
	List(namespace string, selector labels.Selector) (ret []*HuaWeiEventLog, err error)
	Get(namespace, name string) (*HuaWeiEventLog, error)
}

type HuaWeiEventLogController interface {
	Informer() cache.SharedIndexInformer
	Lister() HuaWeiEventLogLister
	AddHandler(name string, handler HuaWeiEventLogHandlerFunc)
	AddClusterScopedHandler(name, clusterName string, handler HuaWeiEventLogHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type HuaWeiEventLogInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*HuaWeiEventLog) (*HuaWeiEventLog, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*HuaWeiEventLog, error)
	Get(name string, opts metav1.GetOptions) (*HuaWeiEventLog, error)
	Update(*HuaWeiEventLog) (*HuaWeiEventLog, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*HuaWeiEventLogList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() HuaWeiEventLogController
	AddHandler(name string, sync HuaWeiEventLogHandlerFunc)
	AddLifecycle(name string, lifecycle HuaWeiEventLogLifecycle)
	AddClusterScopedHandler(name, clusterName string, sync HuaWeiEventLogHandlerFunc)
	AddClusterScopedLifecycle(name, clusterName string, lifecycle HuaWeiEventLogLifecycle)
}

type huaWeiEventLogLister struct {
	controller *huaWeiEventLogController
}

func (l *huaWeiEventLogLister) List(namespace string, selector labels.Selector) (ret []*HuaWeiEventLog, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*HuaWeiEventLog))
	})
	return
}

func (l *huaWeiEventLogLister) Get(namespace, name string) (*HuaWeiEventLog, error) {
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
			Group:    HuaWeiEventLogGroupVersionKind.Group,
			Resource: "huaWeiEventLog",
		}, key)
	}
	return obj.(*HuaWeiEventLog), nil
}

type huaWeiEventLogController struct {
	controller.GenericController
}

func (c *huaWeiEventLogController) Lister() HuaWeiEventLogLister {
	return &huaWeiEventLogLister{
		controller: c,
	}
}

func (c *huaWeiEventLogController) AddHandler(name string, handler HuaWeiEventLogHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*HuaWeiEventLog))
	})
}

func (c *huaWeiEventLogController) AddClusterScopedHandler(name, cluster string, handler HuaWeiEventLogHandlerFunc) {
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

		return handler(key, obj.(*HuaWeiEventLog))
	})
}

type huaWeiEventLogFactory struct {
}

func (c huaWeiEventLogFactory) Object() runtime.Object {
	return &HuaWeiEventLog{}
}

func (c huaWeiEventLogFactory) List() runtime.Object {
	return &HuaWeiEventLogList{}
}

func (s *huaWeiEventLogClient) Controller() HuaWeiEventLogController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.huaWeiEventLogControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(HuaWeiEventLogGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &huaWeiEventLogController{
		GenericController: genericController,
	}

	s.client.huaWeiEventLogControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type huaWeiEventLogClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   HuaWeiEventLogController
}

func (s *huaWeiEventLogClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *huaWeiEventLogClient) Create(o *HuaWeiEventLog) (*HuaWeiEventLog, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*HuaWeiEventLog), err
}

func (s *huaWeiEventLogClient) Get(name string, opts metav1.GetOptions) (*HuaWeiEventLog, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*HuaWeiEventLog), err
}

func (s *huaWeiEventLogClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*HuaWeiEventLog, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*HuaWeiEventLog), err
}

func (s *huaWeiEventLogClient) Update(o *HuaWeiEventLog) (*HuaWeiEventLog, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*HuaWeiEventLog), err
}

func (s *huaWeiEventLogClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *huaWeiEventLogClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *huaWeiEventLogClient) List(opts metav1.ListOptions) (*HuaWeiEventLogList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*HuaWeiEventLogList), err
}

func (s *huaWeiEventLogClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *huaWeiEventLogClient) Patch(o *HuaWeiEventLog, data []byte, subresources ...string) (*HuaWeiEventLog, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*HuaWeiEventLog), err
}

func (s *huaWeiEventLogClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *huaWeiEventLogClient) AddHandler(name string, sync HuaWeiEventLogHandlerFunc) {
	s.Controller().AddHandler(name, sync)
}

func (s *huaWeiEventLogClient) AddLifecycle(name string, lifecycle HuaWeiEventLogLifecycle) {
	sync := NewHuaWeiEventLogLifecycleAdapter(name, false, s, lifecycle)
	s.AddHandler(name, sync)
}

func (s *huaWeiEventLogClient) AddClusterScopedHandler(name, clusterName string, sync HuaWeiEventLogHandlerFunc) {
	s.Controller().AddClusterScopedHandler(name, clusterName, sync)
}

func (s *huaWeiEventLogClient) AddClusterScopedLifecycle(name, clusterName string, lifecycle HuaWeiEventLogLifecycle) {
	sync := NewHuaWeiEventLogLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.AddClusterScopedHandler(name, clusterName, sync)
}
