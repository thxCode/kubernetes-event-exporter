package v3

import (
	"github.com/rancher/norman/types"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HuaWeiClusterEvent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              HuaWeiClusterEventSpec   `json:"spec"`
	Status            HuaWeiClusterEventStatus `json:"status"`
}

type HuaWeiClusterEventSpec struct {
	DisplayName string `json:"displayName,omitempty"`
	ClusterName string `json:"clusterName" norman:"type=reference[cluster]"`
}

type HuaWeiClusterEventStatus struct {
	IsRemoved bool `json:"isRemoved"`
}

type HuaWeiEventLog struct {
	types.Namespaced

	*v1.Event  `json:",inline"`
	EventName  string   `json:"eventName" norman:"type=reference[huaWeiClusterEvent]"`
	AttachPod  *v1.Pod  `json:"attachPod,omitempty"`
	AttachNode *v1.Node `json:"attachNode,omitempty"`
}

type HuaWeiEventLogFilter struct {
	Namespace          string `json:"namespace,omitempty"`
	EventName          string `json:"eventName,omitempty" norman:"required,type=reference[huaWeiClusterEvent]"`
	LogType            string `json:"logType,omitempty" norman:"type=enum,options=All|Warning|Normal,default=All"`
	ResourceKind       string `json:"resourceKind,omitempty" norman:"type=enum,options=All|Pod|Node|Container,default=All"`
	CreatedRangeFormat string `json:"createdRangeFormat,omitempty" norman:"default=2006-01-02T15:04:05Z07:00"`
	CreatedRangeStart  string `json:"createdRangeStart,omitempty"`
	CreatedRangeEnd    string `json:"createdRangeEnd,omitempty"`
}
