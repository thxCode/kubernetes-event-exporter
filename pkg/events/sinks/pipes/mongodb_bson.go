package pipes

import (
	"unsafe"

	"github.com/mongodb/mongo-go-driver/bson"
	apiCoreV1 "k8s.io/api/core/v1"
	apisMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func toBsonDocument(elems ...*bson.Element) *bson.Document {
	doc := bson.NewDocument()
	doc.IgnoreNilInsert = true

	doc.Append(elems...)

	return doc
}

func toBsonSubDocumentFromElements(key string, elems ...*bson.Element) *bson.Element {
	return bson.EC.SubDocument(key, toBsonDocument(elems...))
}

func toBsonInt32Element(key string, value *int32) *bson.Element {
	if value == nil || *value == 0 {
		return nil
	}

	return bson.EC.Int32(key, *value)
}

func toBsonInt64Element(key string, value *int64) *bson.Element {
	if value == nil || *value == 0 {
		return nil
	}

	return bson.EC.Int64(key, *value)
}

func toBsonStringElement(key string, value *string) *bson.Element {
	if value == nil || len(*value) == 0 {
		return nil
	}

	return bson.EC.String(key, *value)
}

func toBsonMetaTimeElement(key string, value *apisMetaV1.Time) *bson.Element {
	if value == nil {
		return nil
	}

	return bson.EC.Time(key, value.Time)
}

func toBsonMetaMicroTimeElement(key string, value *apisMetaV1.MicroTime) *bson.Element {
	if value == nil {
		return nil
	}

	return bson.EC.Time(key, value.Time)
}

func toBsonBoolElement(key string, value *bool) *bson.Element {
	if value == nil {
		return nil
	}

	return bson.EC.Boolean(key, *value)
}

func toBsonString2StringMapElement(key string, value map[string]string) *bson.Element {
	if value == nil {
		return nil
	}

	doc := bson.NewDocument()
	for valueKey, valueValue := range value {
		doc.Append(bson.EC.String(valueKey, valueValue))
	}

	return bson.EC.SubDocument(key, doc)
}

func toBsonStringArrayElement(key string, value []string) *bson.Element {
	if value == nil {
		return nil
	}

	arr := bson.NewArray()
	for _, valueValue := range value {
		arr.Append(bson.VC.String(valueValue))
	}

	return bson.EC.Array(key, arr)
}

func toBsonMetaOwnerReferenceArrayElement(key string, value []apisMetaV1.OwnerReference) *bson.Element {
	if value == nil {
		return nil
	}

	arr := bson.NewArray()
	for _, valueValue := range value {
		arr.Append(bson.VC.Document(toBsonDocument(
			toBsonStringElement("apiVersion", &valueValue.APIVersion),
			toBsonStringElement("kind", &valueValue.Kind),
			toBsonStringElement("name", &valueValue.Name),
			toBsonStringElement("uid", (*string)(unsafe.Pointer(&valueValue.UID))),
			toBsonBoolElement("controller", valueValue.Controller),
			toBsonBoolElement("blockOwnerDeletion", valueValue.BlockOwnerDeletion),
		)))
	}

	return bson.EC.Array(key, arr)
}

func toBsonMetaInitializerArrayElement(key string, value []apisMetaV1.Initializer) *bson.Element {
	if value == nil {
		return nil
	}

	arr := bson.NewArray()
	for _, valueValue := range value {
		arr.Append(bson.VC.Document(bson.NewDocument(
			toBsonStringElement("name", &valueValue.Name),
		)))
	}

	return bson.EC.Array(key, arr)
}

func toBsonMetaStatusCauseArrayElement(key string, value []apisMetaV1.StatusCause) *bson.Element {
	if value == nil {
		return nil
	}

	arr := bson.NewArray()
	for _, valueValue := range value {
		arr.Append(bson.VC.Document(bson.NewDocument(
			toBsonStringElement("type", (*string)(unsafe.Pointer(&valueValue.Type))),
			toBsonStringElement("message", &valueValue.Message),
			toBsonStringElement("Field", &valueValue.Field),
		)))
	}

	return bson.EC.Array(key, arr)
}

func toBsonMetaStatusDetailsElement(key string, value *apisMetaV1.StatusDetails) *bson.Element {
	if value == nil {
		return nil
	}

	return bson.EC.SubDocument(key, toBsonDocument(
		toBsonStringElement("name", &value.Name),
		toBsonStringElement("group", &value.Group),
		toBsonStringElement("kind", &value.Kind),
		toBsonStringElement("uid", (*string)(unsafe.Pointer(&value.UID))),
		toBsonMetaStatusCauseArrayElement("causes", value.Causes),
		toBsonInt32Element("retryAfterSeconds", &value.RetryAfterSeconds),
	))
}

func toBsonMetaStatusElement(key string, value *apisMetaV1.Status) *bson.Element {
	if value == nil {
		return nil
	}

	metadata := value.ListMeta

	return bson.EC.SubDocument(key, toBsonDocument(
		toBsonStringElement("kind", &value.Kind),
		toBsonStringElement("apiVersion", &value.APIVersion),
		toBsonSubDocumentFromElements("metadata",
			toBsonStringElement("selfLink", &metadata.SelfLink),
			toBsonStringElement("resourceVersion", &metadata.ResourceVersion),
			toBsonStringElement("continue", &metadata.Continue),
		),
		toBsonStringElement("status", &value.Status),
		toBsonStringElement("message", &value.Message),
		toBsonStringElement("reason", (*string)(unsafe.Pointer(&value.Reason))),
		toBsonMetaStatusDetailsElement("details", value.Details),
		toBsonInt32Element("code", &value.Code),
	))
}

func toBsonMetaInitializersElement(key string, value *apisMetaV1.Initializers) *bson.Element {
	if value == nil {
		return nil
	}

	return bson.EC.SubDocument(key, toBsonDocument(
		toBsonMetaInitializerArrayElement("pending", value.Pending),
		toBsonMetaStatusElement("result", value.Result),
	))
}

func toBsonMetaObjectMetaElement(key string, value *apisMetaV1.ObjectMeta) *bson.Element {
	if value == nil {
		return nil
	}

	return toBsonSubDocumentFromElements(key,
		toBsonStringElement("name", &value.Name),
		toBsonStringElement("generateName", &value.GenerateName),
		toBsonStringElement("namespace", &value.Namespace),
		toBsonStringElement("selfLink", &value.SelfLink),
		toBsonStringElement("uid", (*string)(unsafe.Pointer(&value.UID))),
		toBsonStringElement("resourceVersion", &value.ResourceVersion),
		toBsonInt64Element("generation", &value.Generation),
		toBsonMetaTimeElement("creationTimestamp", &value.CreationTimestamp),
		toBsonMetaTimeElement("deletionTimestamp", value.DeletionTimestamp),
		toBsonInt64Element("deletionGracePeriodSeconds", value.DeletionGracePeriodSeconds),
		toBsonString2StringMapElement("labels", value.Labels),
		toBsonString2StringMapElement("annotations", value.Annotations),
		toBsonMetaOwnerReferenceArrayElement("ownerReferences", value.OwnerReferences),
		toBsonMetaInitializersElement("initializers", value.Initializers),
		toBsonStringArrayElement("finalizers", value.Finalizers),
		toBsonStringElement("clusterName", &value.ClusterName),
	)
}

func toBsonEventSeriesElement(key string, value *apiCoreV1.EventSeries) *bson.Element {
	if value == nil {
		return nil
	}

	return toBsonSubDocumentFromElements(key,
		toBsonInt32Element("count", &value.Count),
		toBsonMetaMicroTimeElement("lastObservedTime", &value.LastObservedTime),
		toBsonStringElement("state", (*string)(unsafe.Pointer(&value.State))),
	)
}

func toBsonObjectReferenceElement(key string, value *apiCoreV1.ObjectReference) *bson.Element {
	if value == nil {
		return nil
	}

	return toBsonSubDocumentFromElements(key,
		toBsonStringElement("kind", &value.Kind),
		toBsonStringElement("namespace", &value.Namespace),
		toBsonStringElement("name", &value.Name),
		toBsonStringElement("uid", (*string)(unsafe.Pointer(&value.UID))),
		toBsonStringElement("apiVersion", &value.APIVersion),
		toBsonStringElement("resourceVersion", &value.ResourceVersion),
		toBsonStringElement("fieldPath", &value.FieldPath),
	)
}

func toBsonEventSourceElement(key string, value *apiCoreV1.EventSource) *bson.Element {
	if value == nil {
		return nil
	}

	return toBsonSubDocumentFromElements(key,
		toBsonStringElement("component", &value.Component),
		toBsonStringElement("host", &value.Host),
	)
}

func eventToBson(value *apiCoreV1.Event) *bson.Document {
	return toBsonDocument(
		toBsonStringElement("kind", &value.Kind),
		toBsonStringElement("apiVersion", &value.APIVersion),
		toBsonMetaObjectMetaElement("metadata", &value.ObjectMeta),
		toBsonObjectReferenceElement("involvedObject", &value.InvolvedObject),
		toBsonStringElement("reason", &value.Reason),
		toBsonStringElement("message", &value.Message),
		toBsonEventSourceElement("source", &value.Source),
		toBsonMetaTimeElement("firstTimestamp", &value.FirstTimestamp),
		toBsonMetaTimeElement("lastTimestamp", &value.LastTimestamp),
		toBsonInt32Element("count", &value.Count),
		toBsonStringElement("type", &value.Type),
		toBsonMetaMicroTimeElement("eventTime", &value.EventTime),
		toBsonEventSeriesElement("series", value.Series),
		toBsonStringElement("action", &value.Action),
		toBsonObjectReferenceElement("related", value.Related),
		toBsonStringElement("reportingComponent", &value.ReportingController),
		toBsonStringElement("reportingInstance", &value.ReportingInstance),
	)
}
