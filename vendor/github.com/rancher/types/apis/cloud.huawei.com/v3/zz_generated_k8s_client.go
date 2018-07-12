package v3

import (
	"context"
	"sync"

	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"github.com/rancher/norman/restwatch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type Interface interface {
	RESTClient() rest.Interface
	controller.Starter

	HuaWeiClusterEventsGetter
	HuaWeiEventLogsGetter
}

type Client struct {
	sync.Mutex
	restClient rest.Interface
	starters   []controller.Starter

	huaWeiClusterEventControllers map[string]HuaWeiClusterEventController
	huaWeiEventLogControllers     map[string]HuaWeiEventLogController
}

func NewForConfig(config rest.Config) (Interface, error) {
	if config.NegotiatedSerializer == nil {
		configConfig := dynamic.ContentConfig()
		config.NegotiatedSerializer = configConfig.NegotiatedSerializer
	}

	restClient, err := restwatch.UnversionedRESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &Client{
		restClient: restClient,

		huaWeiClusterEventControllers: map[string]HuaWeiClusterEventController{},
		huaWeiEventLogControllers:     map[string]HuaWeiEventLogController{},
	}, nil
}

func (c *Client) RESTClient() rest.Interface {
	return c.restClient
}

func (c *Client) Sync(ctx context.Context) error {
	return controller.Sync(ctx, c.starters...)
}

func (c *Client) Start(ctx context.Context, threadiness int) error {
	return controller.Start(ctx, threadiness, c.starters...)
}

type HuaWeiClusterEventsGetter interface {
	HuaWeiClusterEvents(namespace string) HuaWeiClusterEventInterface
}

func (c *Client) HuaWeiClusterEvents(namespace string) HuaWeiClusterEventInterface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &HuaWeiClusterEventResource, HuaWeiClusterEventGroupVersionKind, huaWeiClusterEventFactory{})
	return &huaWeiClusterEventClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type HuaWeiEventLogsGetter interface {
	HuaWeiEventLogs(namespace string) HuaWeiEventLogInterface
}

func (c *Client) HuaWeiEventLogs(namespace string) HuaWeiEventLogInterface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &HuaWeiEventLogResource, HuaWeiEventLogGroupVersionKind, huaWeiEventLogFactory{})
	return &huaWeiEventLogClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}
