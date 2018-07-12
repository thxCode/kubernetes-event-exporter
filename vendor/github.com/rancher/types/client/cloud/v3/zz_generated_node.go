package client

import (
	"github.com/rancher/norman/types"
)

const (
	NodeType                      = "node"
	NodeFieldAnnotations          = "annotations"
	NodeFieldConfigSource         = "configSource"
	NodeFieldCreated              = "created"
	NodeFieldCreatorID            = "creatorId"
	NodeFieldExternalID           = "externalID"
	NodeFieldLabels               = "labels"
	NodeFieldName                 = "name"
	NodeFieldOwnerReferences      = "ownerReferences"
	NodeFieldPodCIDR              = "podCIDR"
	NodeFieldProviderID           = "providerID"
	NodeFieldRemoved              = "removed"
	NodeFieldState                = "state"
	NodeFieldStatus               = "status"
	NodeFieldTaints               = "taints"
	NodeFieldTransitioning        = "transitioning"
	NodeFieldTransitioningMessage = "transitioningMessage"
	NodeFieldUnschedulable        = "unschedulable"
	NodeFieldUuid                 = "uuid"
)

type Node struct {
	types.Resource
	Annotations          map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	ConfigSource         *NodeConfigSource `json:"configSource,omitempty" yaml:"configSource,omitempty"`
	Created              string            `json:"created,omitempty" yaml:"created,omitempty"`
	CreatorID            string            `json:"creatorId,omitempty" yaml:"creatorId,omitempty"`
	ExternalID           string            `json:"externalID,omitempty" yaml:"externalID,omitempty"`
	Labels               map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Name                 string            `json:"name,omitempty" yaml:"name,omitempty"`
	OwnerReferences      []OwnerReference  `json:"ownerReferences,omitempty" yaml:"ownerReferences,omitempty"`
	PodCIDR              string            `json:"podCIDR,omitempty" yaml:"podCIDR,omitempty"`
	ProviderID           string            `json:"providerID,omitempty" yaml:"providerID,omitempty"`
	Removed              string            `json:"removed,omitempty" yaml:"removed,omitempty"`
	State                string            `json:"state,omitempty" yaml:"state,omitempty"`
	Status               *NodeStatus       `json:"status,omitempty" yaml:"status,omitempty"`
	Taints               []Taint           `json:"taints,omitempty" yaml:"taints,omitempty"`
	Transitioning        string            `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`
	TransitioningMessage string            `json:"transitioningMessage,omitempty" yaml:"transitioningMessage,omitempty"`
	Unschedulable        bool              `json:"unschedulable,omitempty" yaml:"unschedulable,omitempty"`
	Uuid                 string            `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}
type NodeCollection struct {
	types.Collection
	Data   []Node `json:"data,omitempty"`
	client *NodeClient
}

type NodeClient struct {
	apiClient *Client
}

type NodeOperations interface {
	List(opts *types.ListOpts) (*NodeCollection, error)
	Create(opts *Node) (*Node, error)
	Update(existing *Node, updates interface{}) (*Node, error)
	ByID(id string) (*Node, error)
	Delete(container *Node) error
}

func newNodeClient(apiClient *Client) *NodeClient {
	return &NodeClient{
		apiClient: apiClient,
	}
}

func (c *NodeClient) Create(container *Node) (*Node, error) {
	resp := &Node{}
	err := c.apiClient.Ops.DoCreate(NodeType, container, resp)
	return resp, err
}

func (c *NodeClient) Update(existing *Node, updates interface{}) (*Node, error) {
	resp := &Node{}
	err := c.apiClient.Ops.DoUpdate(NodeType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *NodeClient) List(opts *types.ListOpts) (*NodeCollection, error) {
	resp := &NodeCollection{}
	err := c.apiClient.Ops.DoList(NodeType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *NodeCollection) Next() (*NodeCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &NodeCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *NodeClient) ByID(id string) (*Node, error) {
	resp := &Node{}
	err := c.apiClient.Ops.DoByID(NodeType, id, resp)
	return resp, err
}

func (c *NodeClient) Delete(container *Node) error {
	return c.apiClient.Ops.DoResourceDelete(NodeType, &container.Resource)
}
