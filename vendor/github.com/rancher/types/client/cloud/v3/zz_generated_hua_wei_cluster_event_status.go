package client

const (
	HuaWeiClusterEventStatusType           = "huaWeiClusterEventStatus"
	HuaWeiClusterEventStatusFieldIsRemoved = "isRemoved"
)

type HuaWeiClusterEventStatus struct {
	IsRemoved bool `json:"isRemoved,omitempty" yaml:"isRemoved,omitempty"`
}
