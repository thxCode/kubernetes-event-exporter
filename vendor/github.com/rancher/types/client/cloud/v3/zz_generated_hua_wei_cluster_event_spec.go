package client

const (
	HuaWeiClusterEventSpecType             = "huaWeiClusterEventSpec"
	HuaWeiClusterEventSpecFieldClusterId   = "clusterId"
	HuaWeiClusterEventSpecFieldDisplayName = "displayName"
)

type HuaWeiClusterEventSpec struct {
	ClusterId   string `json:"clusterId,omitempty" yaml:"clusterId,omitempty"`
	DisplayName string `json:"displayName,omitempty" yaml:"displayName,omitempty"`
}
