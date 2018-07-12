package client

const (
	ObjectReferenceType              = "objectReference"
	ObjectReferenceFieldFieldPath    = "fieldPath"
	ObjectReferenceFieldNamespace    = "namespace"
	ObjectReferenceFieldResourceKind = "resourceKind"
	ObjectReferenceFieldResourceName = "resourceName"
)

type ObjectReference struct {
	FieldPath    string `json:"fieldPath,omitempty" yaml:"fieldPath,omitempty"`
	Namespace    string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	ResourceKind string `json:"resourceKind,omitempty" yaml:"resourceKind,omitempty"`
	ResourceName string `json:"resourceName,omitempty" yaml:"resourceName,omitempty"`
}
