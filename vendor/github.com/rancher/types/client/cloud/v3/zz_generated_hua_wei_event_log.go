package client

import (
	"github.com/rancher/norman/types"
)

const (
	HuaWeiEventLogType                     = "huaWeiEventLog"
	HuaWeiEventLogFieldAction              = "action"
	HuaWeiEventLogFieldAnnotations         = "annotations"
	HuaWeiEventLogFieldAttachNode          = "attachNode"
	HuaWeiEventLogFieldAttachPod           = "attachPod"
	HuaWeiEventLogFieldCount               = "count"
	HuaWeiEventLogFieldCreated             = "created"
	HuaWeiEventLogFieldEventId             = "eventId"
	HuaWeiEventLogFieldEventTime           = "eventTime"
	HuaWeiEventLogFieldFieldPath           = "fieldPath"
	HuaWeiEventLogFieldFirstTimestamp      = "firstTimestamp"
	HuaWeiEventLogFieldLabels              = "labels"
	HuaWeiEventLogFieldLastTimestamp       = "lastTimestamp"
	HuaWeiEventLogFieldLogType             = "logType"
	HuaWeiEventLogFieldMessage             = "message"
	HuaWeiEventLogFieldName                = "name"
	HuaWeiEventLogFieldNamespace           = "namespace"
	HuaWeiEventLogFieldNamespaceId         = "namespaceId"
	HuaWeiEventLogFieldOwnerReferences     = "ownerReferences"
	HuaWeiEventLogFieldReason              = "reason"
	HuaWeiEventLogFieldRelated             = "related"
	HuaWeiEventLogFieldRemoved             = "removed"
	HuaWeiEventLogFieldReportingController = "reportingComponent"
	HuaWeiEventLogFieldReportingInstance   = "reportingInstance"
	HuaWeiEventLogFieldResourceKind        = "resourceKind"
	HuaWeiEventLogFieldResourceName        = "resourceName"
	HuaWeiEventLogFieldSeries              = "series"
	HuaWeiEventLogFieldSource              = "source"
	HuaWeiEventLogFieldUuid                = "uuid"
)

type HuaWeiEventLog struct {
	types.Resource
	Action              string            `json:"action,omitempty" yaml:"action,omitempty"`
	Annotations         map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	AttachNode          *Node             `json:"attachNode,omitempty" yaml:"attachNode,omitempty"`
	AttachPod           *Pod              `json:"attachPod,omitempty" yaml:"attachPod,omitempty"`
	Count               int64             `json:"count,omitempty" yaml:"count,omitempty"`
	Created             string            `json:"created,omitempty" yaml:"created,omitempty"`
	EventId             string            `json:"eventId,omitempty" yaml:"eventId,omitempty"`
	EventTime           *MicroTime        `json:"eventTime,omitempty" yaml:"eventTime,omitempty"`
	FieldPath           string            `json:"fieldPath,omitempty" yaml:"fieldPath,omitempty"`
	FirstTimestamp      string            `json:"firstTimestamp,omitempty" yaml:"firstTimestamp,omitempty"`
	Labels              map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	LastTimestamp       string            `json:"lastTimestamp,omitempty" yaml:"lastTimestamp,omitempty"`
	LogType             string            `json:"logType,omitempty" yaml:"logType,omitempty"`
	Message             string            `json:"message,omitempty" yaml:"message,omitempty"`
	Name                string            `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace           string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	NamespaceId         string            `json:"namespaceId,omitempty" yaml:"namespaceId,omitempty"`
	OwnerReferences     []OwnerReference  `json:"ownerReferences,omitempty" yaml:"ownerReferences,omitempty"`
	Reason              string            `json:"reason,omitempty" yaml:"reason,omitempty"`
	Related             *ObjectReference  `json:"related,omitempty" yaml:"related,omitempty"`
	Removed             string            `json:"removed,omitempty" yaml:"removed,omitempty"`
	ReportingController string            `json:"reportingComponent,omitempty" yaml:"reportingComponent,omitempty"`
	ReportingInstance   string            `json:"reportingInstance,omitempty" yaml:"reportingInstance,omitempty"`
	ResourceKind        string            `json:"resourceKind,omitempty" yaml:"resourceKind,omitempty"`
	ResourceName        string            `json:"resourceName,omitempty" yaml:"resourceName,omitempty"`
	Series              *EventSeries      `json:"series,omitempty" yaml:"series,omitempty"`
	Source              *EventSource      `json:"source,omitempty" yaml:"source,omitempty"`
	Uuid                string            `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}
type HuaWeiEventLogCollection struct {
	types.Collection
	Data   []HuaWeiEventLog `json:"data,omitempty"`
	client *HuaWeiEventLogClient
}

type HuaWeiEventLogClient struct {
	apiClient *Client
}

type HuaWeiEventLogOperations interface {
	List(opts *types.ListOpts) (*HuaWeiEventLogCollection, error)
	Create(opts *HuaWeiEventLog) (*HuaWeiEventLog, error)
	Update(existing *HuaWeiEventLog, updates interface{}) (*HuaWeiEventLog, error)
	ByID(id string) (*HuaWeiEventLog, error)
	Delete(container *HuaWeiEventLog) error
}

func newHuaWeiEventLogClient(apiClient *Client) *HuaWeiEventLogClient {
	return &HuaWeiEventLogClient{
		apiClient: apiClient,
	}
}

func (c *HuaWeiEventLogClient) Create(container *HuaWeiEventLog) (*HuaWeiEventLog, error) {
	resp := &HuaWeiEventLog{}
	err := c.apiClient.Ops.DoCreate(HuaWeiEventLogType, container, resp)
	return resp, err
}

func (c *HuaWeiEventLogClient) Update(existing *HuaWeiEventLog, updates interface{}) (*HuaWeiEventLog, error) {
	resp := &HuaWeiEventLog{}
	err := c.apiClient.Ops.DoUpdate(HuaWeiEventLogType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *HuaWeiEventLogClient) List(opts *types.ListOpts) (*HuaWeiEventLogCollection, error) {
	resp := &HuaWeiEventLogCollection{}
	err := c.apiClient.Ops.DoList(HuaWeiEventLogType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *HuaWeiEventLogCollection) Next() (*HuaWeiEventLogCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &HuaWeiEventLogCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *HuaWeiEventLogClient) ByID(id string) (*HuaWeiEventLog, error) {
	resp := &HuaWeiEventLog{}
	err := c.apiClient.Ops.DoByID(HuaWeiEventLogType, id, resp)
	return resp, err
}

func (c *HuaWeiEventLogClient) Delete(container *HuaWeiEventLog) error {
	return c.apiClient.Ops.DoResourceDelete(HuaWeiEventLogType, &container.Resource)
}
