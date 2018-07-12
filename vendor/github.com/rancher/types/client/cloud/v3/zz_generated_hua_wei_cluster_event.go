package client

import (
	"github.com/rancher/norman/types"
)

const (
	HuaWeiClusterEventType                      = "huaWeiClusterEvent"
	HuaWeiClusterEventFieldAnnotations          = "annotations"
	HuaWeiClusterEventFieldClusterId            = "clusterId"
	HuaWeiClusterEventFieldCreated              = "created"
	HuaWeiClusterEventFieldIsRemoved            = "isRemoved"
	HuaWeiClusterEventFieldLabels               = "labels"
	HuaWeiClusterEventFieldName                 = "name"
	HuaWeiClusterEventFieldOwnerReferences      = "ownerReferences"
	HuaWeiClusterEventFieldRemoved              = "removed"
	HuaWeiClusterEventFieldState                = "state"
	HuaWeiClusterEventFieldTransitioning        = "transitioning"
	HuaWeiClusterEventFieldTransitioningMessage = "transitioningMessage"
	HuaWeiClusterEventFieldUuid                 = "uuid"
)

type HuaWeiClusterEvent struct {
	types.Resource
	Annotations          map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	ClusterId            string            `json:"clusterId,omitempty" yaml:"clusterId,omitempty"`
	Created              string            `json:"created,omitempty" yaml:"created,omitempty"`
	IsRemoved            bool              `json:"isRemoved,omitempty" yaml:"isRemoved,omitempty"`
	Labels               map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Name                 string            `json:"name,omitempty" yaml:"name,omitempty"`
	OwnerReferences      []OwnerReference  `json:"ownerReferences,omitempty" yaml:"ownerReferences,omitempty"`
	Removed              string            `json:"removed,omitempty" yaml:"removed,omitempty"`
	State                string            `json:"state,omitempty" yaml:"state,omitempty"`
	Transitioning        string            `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`
	TransitioningMessage string            `json:"transitioningMessage,omitempty" yaml:"transitioningMessage,omitempty"`
	Uuid                 string            `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}
type HuaWeiClusterEventCollection struct {
	types.Collection
	Data   []HuaWeiClusterEvent `json:"data,omitempty"`
	client *HuaWeiClusterEventClient
}

type HuaWeiClusterEventClient struct {
	apiClient *Client
}

type HuaWeiClusterEventOperations interface {
	List(opts *types.ListOpts) (*HuaWeiClusterEventCollection, error)
	Create(opts *HuaWeiClusterEvent) (*HuaWeiClusterEvent, error)
	Update(existing *HuaWeiClusterEvent, updates interface{}) (*HuaWeiClusterEvent, error)
	ByID(id string) (*HuaWeiClusterEvent, error)
	Delete(container *HuaWeiClusterEvent) error
}

func newHuaWeiClusterEventClient(apiClient *Client) *HuaWeiClusterEventClient {
	return &HuaWeiClusterEventClient{
		apiClient: apiClient,
	}
}

func (c *HuaWeiClusterEventClient) Create(container *HuaWeiClusterEvent) (*HuaWeiClusterEvent, error) {
	resp := &HuaWeiClusterEvent{}
	err := c.apiClient.Ops.DoCreate(HuaWeiClusterEventType, container, resp)
	return resp, err
}

func (c *HuaWeiClusterEventClient) Update(existing *HuaWeiClusterEvent, updates interface{}) (*HuaWeiClusterEvent, error) {
	resp := &HuaWeiClusterEvent{}
	err := c.apiClient.Ops.DoUpdate(HuaWeiClusterEventType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *HuaWeiClusterEventClient) List(opts *types.ListOpts) (*HuaWeiClusterEventCollection, error) {
	resp := &HuaWeiClusterEventCollection{}
	err := c.apiClient.Ops.DoList(HuaWeiClusterEventType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *HuaWeiClusterEventCollection) Next() (*HuaWeiClusterEventCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &HuaWeiClusterEventCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *HuaWeiClusterEventClient) ByID(id string) (*HuaWeiClusterEvent, error) {
	resp := &HuaWeiClusterEvent{}
	err := c.apiClient.Ops.DoByID(HuaWeiClusterEventType, id, resp)
	return resp, err
}

func (c *HuaWeiClusterEventClient) Delete(container *HuaWeiClusterEvent) error {
	return c.apiClient.Ops.DoResourceDelete(HuaWeiClusterEventType, &container.Resource)
}
