package client

const (
	HuaWeiEventLogFilterType                    = "huaWeiEventLogFilter"
	HuaWeiEventLogFilterFieldCreatedRangeEnd    = "createdRangeEnd"
	HuaWeiEventLogFilterFieldCreatedRangeFormat = "createdRangeFormat"
	HuaWeiEventLogFilterFieldCreatedRangeStart  = "createdRangeStart"
	HuaWeiEventLogFilterFieldEventId            = "eventId"
	HuaWeiEventLogFilterFieldLogType            = "logType"
	HuaWeiEventLogFilterFieldNamespaceId        = "namespaceId"
	HuaWeiEventLogFilterFieldResourceKind       = "resourceKind"
)

type HuaWeiEventLogFilter struct {
	CreatedRangeEnd    string `json:"createdRangeEnd,omitempty" yaml:"createdRangeEnd,omitempty"`
	CreatedRangeFormat string `json:"createdRangeFormat,omitempty" yaml:"createdRangeFormat,omitempty"`
	CreatedRangeStart  string `json:"createdRangeStart,omitempty" yaml:"createdRangeStart,omitempty"`
	EventId            string `json:"eventId,omitempty" yaml:"eventId,omitempty"`
	LogType            string `json:"logType,omitempty" yaml:"logType,omitempty"`
	NamespaceId        string `json:"namespaceId,omitempty" yaml:"namespaceId,omitempty"`
	ResourceKind       string `json:"resourceKind,omitempty" yaml:"resourceKind,omitempty"`
}
