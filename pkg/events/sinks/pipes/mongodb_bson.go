package pipes

import (
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/mongodb/mongo-go-driver/bson"
	typesHuawei "github.com/rancher/types/apis/cloud.huawei.com/v3"
	apiCore "k8s.io/api/core/v1"
	apisMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func attachBsonDocumentElement(key string, doc *bson.Document) *bson.Element {
	if doc == nil || doc.Len() == 0 {
		return nil
	}

	return bson.EC.SubDocument(key, doc)
}

func attachBsonArrayElement(key string, array *bson.Array) *bson.Element {
	if array == nil || array.Len() == 0 {
		return nil
	}

	return bson.EC.Array(key, array)
}

func toBsonMapElement(key string, m interface{}) *bson.Element {
	if m == nil {
		return nil
	}

	doc := bson.NewDocument()
	switch m.(type) {
	case map[string]string:
		for mkey, mvalue := range m.(map[string]string) {
			doc.Append(bson.EC.String(mkey, mvalue))
		}
	}

	return attachBsonDocumentElement(key, doc)
}

func toBsonArrayElement(key string, slice interface{}) *bson.Element {
	if slice == nil {
		return nil
	}

	arr := bson.NewArray()
	switch slice.(type) {
	case []string:
		for _, value := range slice.([]string) {
			arr.Append(bson.VC.String(value))
		}
	}

	return attachBsonArrayElement(key, arr)
}

func toBsonElement(key string, value interface{}) *bson.Element {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case *bool:
		return bson.EC.Boolean(key, *v)
	case *string:
		return bson.EC.String(key, *v)
	case *int32:
		return bson.EC.Int32(key, *v)
	case *int64:
		return bson.EC.Int64(key, *v)
	case *apisMeta.Time:
		return bson.EC.Time(key, v.Time)
	case *apisMeta.MicroTime:
		return bson.EC.Time(key, v.Time)
	}

	return nil
}

func newIgnoreNilInsertBsonDocument() *bson.Document {
	ret := bson.NewDocument()
	ret.IgnoreNilInsert = true

	return ret
}

type HuaWeiEventLogBson typesHuawei.HuaWeiEventLog

func (o *HuaWeiEventLogBson) MarshalBSONDocument() (*bson.Document, error) {
	if o == nil {
		return nil, nil
	}

	metadataDoc := ApisMetaObjectMataBson(o.ObjectMeta).MustMarshalBSONDocument()
	sourceDoc := ApiCoreEventSourceBson(o.Source).MustMarshalBSONDocument()
	involvedObjectDoc := ApiCoreObjectReferenceBson(o.InvolvedObject).MustMarshalBSONDocument()
	var relatedDoc *bson.Document
	if o.Related != nil {
		relatedDoc = ApiCoreObjectReferenceBson(*o.Related).MustMarshalBSONDocument()
	}
	var seriesDoc *bson.Document
	if o.Series != nil {
		seriesDoc = ApiCoreEventSeriesBson(*o.Series).MustMarshalBSONDocument()
	}

	doc := newIgnoreNilInsertBsonDocument()
	doc.Append(
		toBsonElement("kind", &o.TypeMeta.Kind),
		toBsonElement("apiVersion", &o.TypeMeta.APIVersion),
		attachBsonDocumentElement("metadata", metadataDoc),
		attachBsonDocumentElement("involvedObject", involvedObjectDoc),
		toBsonElement("reason", &o.Reason),
		toBsonElement("message", &o.Message),
		attachBsonDocumentElement("source", sourceDoc),
		toBsonElement("firstTimestamp", &o.FirstTimestamp),
		toBsonElement("lastTimestamp", &o.LastTimestamp),
		toBsonElement("count", &o.Count),
		toBsonElement("type", &o.Type),
		toBsonElement("eventTime", &o.EventTime),
		attachBsonDocumentElement("series", seriesDoc),
		toBsonElement("action", &o.Action),
		attachBsonDocumentElement("related", relatedDoc),
		toBsonElement("reportingComponent", &o.ReportingController),
		toBsonElement("reportingInstance", &o.ReportingInstance),
	)

	return doc, nil
}

func (o HuaWeiEventLogBson) MustMarshalBSONDocument() *bson.Document {
	doc, _ := o.MarshalBSONDocument()

	return doc
}

type ApisMetaObjectMataBson apisMeta.ObjectMeta

func (o ApisMetaObjectMataBson) MarshalBSONDocument() (*bson.Document, error) {
	var ownerReferenceArray *bson.Array
	if o.OwnerReferences != nil {
		ownerReferenceArray = bson.NewArray()

		for _, ownerReference := range o.OwnerReferences {
			ownerReferenceArray.Append(
				bson.VC.Document(ApisMetaOwnerReferenceBson(ownerReference).MustMarshalBSONDocument()),
			)
		}
	}

	var initializersDoc *bson.Document
	if o.Initializers != nil {
		initializersDoc = ApisMetaInitializersBson(*o.Initializers).MustMarshalBSONDocument()
	}

	doc := newIgnoreNilInsertBsonDocument()
	doc.Append(
		toBsonElement("name", &o.Name),
		toBsonElement("generateName", &o.GenerateName),
		toBsonElement("namespace", &o.Namespace),
		toBsonElement("selfLink", &o.SelfLink),
		toBsonElement("uid", (*string)(unsafe.Pointer(&o.UID))),
		toBsonElement("resourceVersion", &o.ResourceVersion),
		toBsonElement("generation", &o.Generation),
		toBsonElement("creationTimestamp", &o.CreationTimestamp),
		toBsonElement("deletionTimestamp", o.DeletionTimestamp),
		toBsonElement("deletionGracePeriodSeconds", o.DeletionGracePeriodSeconds),
		toBsonMapElement("labels", o.Labels),
		toBsonMapElement("annotations", o.Annotations),
		attachBsonArrayElement("ownerReferences", ownerReferenceArray),
		attachBsonDocumentElement("initializers", initializersDoc),
		toBsonArrayElement("finalizers", o.Finalizers),
		toBsonElement("clusterName", &o.ClusterName),
	)

	if doc.Len() == 0 {
		return nil, nil
	}

	return doc, nil
}

func (o ApisMetaObjectMataBson) MustMarshalBSONDocument() *bson.Document {
	doc, err := o.MarshalBSONDocument()
	if err != nil {
		return nil
	}

	return doc
}

type ApisMetaOwnerReferenceBson apisMeta.OwnerReference

func (o ApisMetaOwnerReferenceBson) MarshalBSONDocument() (*bson.Document, error) {
	doc := newIgnoreNilInsertBsonDocument()
	doc.Append(
		toBsonElement("apiVersion", &o.APIVersion),
		toBsonElement("kind", &o.Kind),
		toBsonElement("name", &o.Name),
		toBsonElement("uid", (*string)(unsafe.Pointer(&o.UID))),
		toBsonElement("controller", o.Controller),
		toBsonElement("blockOwnerDeletion", o.BlockOwnerDeletion),
	)

	if doc.Len() == 0 {
		return nil, nil
	}

	return doc, nil
}

func (o ApisMetaOwnerReferenceBson) MustMarshalBSONDocument() *bson.Document {
	doc, err := o.MarshalBSONDocument()
	if err != nil {
		return nil
	}

	return doc
}

type ApisMetaInitializersBson apisMeta.Initializers

func (o ApisMetaInitializersBson) MarshalBSONDocument() (*bson.Document, error) {
	var pendingArray *bson.Array
	if o.Pending != nil {
		pendingArray = bson.NewArray()

		for _, item := range o.Pending {
			pendingArray.Append(
				bson.VC.Document(ApisMetaInitializerBson(item).MustMarshalBSONDocument()),
			)
		}
	}

	var resultDoc *bson.Document
	if o.Result != nil {
		resultDoc = ApisMetaStatusBson(*o.Result).MustMarshalBSONDocument()
	}

	doc := newIgnoreNilInsertBsonDocument()
	doc.Append(
		attachBsonArrayElement("pending", pendingArray),
		attachBsonDocumentElement("result", resultDoc),
	)

	if doc.Len() == 0 {
		return nil, nil
	}

	return doc, nil
}

func (o ApisMetaInitializersBson) MustMarshalBSONDocument() *bson.Document {
	doc, err := o.MarshalBSONDocument()
	if err != nil {
		return nil
	}

	return doc
}

type ApisMetaInitializerBson apisMeta.Initializer

func (o ApisMetaInitializerBson) MarshalBSONDocument() (*bson.Document, error) {
	doc := newIgnoreNilInsertBsonDocument()
	doc.Append(
		toBsonElement("name", &o.Name),
	)

	if doc.Len() == 0 {
		return nil, nil
	}

	return doc, nil
}

func (o ApisMetaInitializerBson) MustMarshalBSONDocument() *bson.Document {
	doc, err := o.MarshalBSONDocument()
	if err != nil {
		return nil
	}

	return doc
}

type ApisMetaStatusBson apisMeta.Status

func (o ApisMetaStatusBson) MarshalBSONDocument() (*bson.Document, error) {
	var detailsDoc *bson.Document
	if o.Details != nil {
		detailsDoc = ApisMetaStatusDetailsBson(*o.Details).MustMarshalBSONDocument()
	}

	metadataDoc := ApisMetaListMetaBson(o.ListMeta).MustMarshalBSONDocument()

	doc := newIgnoreNilInsertBsonDocument()
	doc.Append(
		toBsonElement("kind", &o.Kind),
		toBsonElement("apiVersion", &o.APIVersion),
		attachBsonDocumentElement("metadata", metadataDoc),
		toBsonElement("status", &o.Status),
		toBsonElement("message", &o.Message),
		toBsonElement("reason", (*string)(unsafe.Pointer(&o.Reason))),
		attachBsonDocumentElement("details", detailsDoc),
		toBsonElement("code", &o.Code),
	)

	if doc.Len() == 0 {
		return nil, nil
	}

	return doc, nil
}

func (o ApisMetaStatusBson) MustMarshalBSONDocument() *bson.Document {
	doc, err := o.MarshalBSONDocument()
	if err != nil {
		return nil
	}

	return doc
}

type ApisMetaListMetaBson apisMeta.ListMeta

func (o ApisMetaListMetaBson) MarshalBSONDocument() (*bson.Document, error) {
	doc := newIgnoreNilInsertBsonDocument()
	doc.Append(
		toBsonElement("selfLink", &o.SelfLink),
		toBsonElement("resourceVersion", &o.ResourceVersion),
		toBsonElement("continue", &o.Continue),
	)

	if doc.Len() == 0 {
		return nil, nil
	}

	return doc, nil
}

func (o ApisMetaListMetaBson) MustMarshalBSONDocument() *bson.Document {
	doc, err := o.MarshalBSONDocument()
	if err != nil {
		return nil
	}

	return doc
}

type ApisMetaStatusDetailsBson apisMeta.StatusDetails

func (o ApisMetaStatusDetailsBson) MarshalBSONDocument() (*bson.Document, error) {
	var causesArray *bson.Array
	if len(o.Causes) != 0 {
		causesArray := bson.NewArray()
		for _, cause := range o.Causes {
			causesArray.Append(
				bson.VC.Document(ApisMetaStatusCauseBson(cause).MustMarshalBSONDocument()),
			)
		}
	}

	doc := newIgnoreNilInsertBsonDocument()
	doc.Append(
		toBsonElement("name", &o.Name),
		toBsonElement("group", &o.Group),
		toBsonElement("kind", &o.Kind),
		toBsonElement("uid", (*string)(unsafe.Pointer(&o.UID))),
		attachBsonArrayElement("causes", causesArray),
		toBsonElement("retryAfterSeconds", &o.RetryAfterSeconds),
	)

	if doc.Len() == 0 {
		return nil, nil
	}

	return doc, nil
}

func (o ApisMetaStatusDetailsBson) MustMarshalBSONDocument() *bson.Document {
	doc, err := o.MarshalBSONDocument()
	if err != nil {
		return nil
	}

	return doc
}

type ApisMetaStatusCauseBson apisMeta.StatusCause

func (o ApisMetaStatusCauseBson) MarshalBSONDocument() (*bson.Document, error) {
	doc := newIgnoreNilInsertBsonDocument()
	doc.Append(
		toBsonElement("type", (*string)(unsafe.Pointer(&o.Type))),
		toBsonElement("message", &o.Message),
		toBsonElement("Field", &o.Field),
	)

	if doc.Len() == 0 {
		return nil, nil
	}

	return doc, nil
}

func (o ApisMetaStatusCauseBson) MustMarshalBSONDocument() *bson.Document {
	doc, err := o.MarshalBSONDocument()
	if err != nil {
		return nil
	}

	return doc
}

type ApiCoreEventSeriesBson apiCore.EventSeries

func (o ApiCoreEventSeriesBson) MarshalBSONDocument() (*bson.Document, error) {
	doc := newIgnoreNilInsertBsonDocument()
	doc.Append(
		toBsonElement("count", &o.Count),
		toBsonElement("lastObservedTime", &o.LastObservedTime),
		toBsonElement("state", (*string)(unsafe.Pointer(&o.State))),
	)

	return doc, nil
}

func (o ApiCoreEventSeriesBson) MustMarshalBSONDocument() *bson.Document {
	doc, _ := o.MarshalBSONDocument()

	return doc
}

type ApiCoreEventSourceBson apiCore.EventSource

func (o ApiCoreEventSourceBson) MarshalBSONDocument() (*bson.Document, error) {
	doc := newIgnoreNilInsertBsonDocument()
	doc.Append(
		toBsonElement("component", &o.Component),
		toBsonElement("host", &o.Host),
	)

	return doc, nil
}

func (o ApiCoreEventSourceBson) MustMarshalBSONDocument() *bson.Document {
	doc, _ := o.MarshalBSONDocument()

	return doc
}

type ApiCoreObjectReferenceBson apiCore.ObjectReference

func (o ApiCoreObjectReferenceBson) MarshalBSONDocument() (*bson.Document, error) {
	doc := newIgnoreNilInsertBsonDocument()
	doc.Append(
		toBsonElement("kind", &o.Kind),
		toBsonElement("namespace", &o.Namespace),
		toBsonElement("name", &o.Name),
		toBsonElement("uid", (*string)(unsafe.Pointer(&o.UID))),
		toBsonElement("apiVersion", &o.APIVersion),
		toBsonElement("resourceVersion", &o.ResourceVersion),
		toBsonElement("fieldPath", &o.FieldPath),
	)

	return doc, nil
}

func (o ApiCoreObjectReferenceBson) MustMarshalBSONDocument() *bson.Document {
	doc, _ := o.MarshalBSONDocument()

	return doc
}

type HuaWeiEventLogBson2 typesHuawei.HuaWeiEventLog

func (o *HuaWeiEventLogBson2) MarshalBSONDocument() (*bson.Document, error) {
	ref := reflect.ValueOf(*o)
	refType := ref.Type()
	root := newIgnoreNilInsertBsonDocument()

	for i := 0; i < ref.NumField(); i++ {
		refItemField := refType.Field(i)
		if len(refItemField.PkgPath) != 0 {
			continue
		}

		if refItemFieldJsonTag := refItemField.Tag.Get("json"); len(refItemFieldJsonTag) != 0 {
			key := refItemFieldJsonTag
			if strings.HasSuffix(refItemFieldJsonTag, ",inline") {
				key = ""
			} else if strings.HasSuffix(refItemFieldJsonTag, ",omitempty") {
				splits := strings.Split(refItemFieldJsonTag, ",")
				key = splits[0]
			}

			root.Append(bsonElement(refItemField.Type.Kind(), key, ref.Field(i))...)
		}
	}

	return root, nil
}

func bsonElement(kind reflect.Kind, key string, value reflect.Value) []*bson.Element {
	switch kind {
	case reflect.Uint32, reflect.Int32:
		return []*bson.Element{bson.EC.Int32(key, int32(value.Int()))}
	case reflect.Uint64, reflect.Int64:
		return []*bson.Element{bson.EC.Int64(key, value.Int())}
	case reflect.String:
		return []*bson.Element{bson.EC.String(key, value.String())}
	case reflect.Bool:
		return []*bson.Element{bson.EC.Boolean(key, value.Bool())}
	case reflect.Float32, reflect.Float64:
		return []*bson.Element{bson.EC.Double(key, value.Float())}
	case reflect.Map:
		if value.IsNil() {
			return nil
		}

		subRoot := newIgnoreNilInsertBsonDocument()

		for _, key := range value.MapKeys() {
			valueValue := value.MapIndex(key)
			subRoot.Append(bsonElement(valueValue.Kind(), key.String(), valueValue)...)
		}

		return []*bson.Element{bson.EC.SubDocument(key, subRoot)}
	case reflect.Slice:
		if value.IsNil() {
			return nil
		}

		subRoot := bson.ArrayFromDocument(newIgnoreNilInsertBsonDocument())
		for i := 0; i < value.Len(); i++ {
			valueValue := value.Index(i)
			subRoot.Append(
				bsonValue(valueValue.Kind(), valueValue),
			)
		}

		return []*bson.Element{bson.EC.Array(key, subRoot)}
	case reflect.Interface:
		if value.IsNil() {
			return nil
		}

		return bsonElement(value.Kind(), key, value.Elem())
	case reflect.Ptr:
		value := value.Elem()

		if !value.IsValid() {
			return nil
		}

		return bsonElement(value.Kind(), key, value)
	case reflect.Struct:
		switch t := value.Interface().(type) {
		case apisMeta.Time:
			return []*bson.Element{bson.EC.DateTime(key, t.Time.Unix())}
		case apisMeta.MicroTime:
			return []*bson.Element{bson.EC.DateTime(key, t.Time.Unix())}
		case time.Time:
			return []*bson.Element{bson.EC.DateTime(key, t.Unix())}
		case apisMeta.Timestamp:
			return []*bson.Element{bson.EC.DateTime(key, time.Unix(t.Seconds, int64(t.Nanos)).Unix())}
		}

		ret := make([]*bson.Element, 0, 16)
		valueType := value.Type()

		for i := 0; i < value.NumField(); i++ {
			valueItemField := valueType.Field(i)
			if len(valueItemField.PkgPath) != 0 {
				continue
			}

			if valueItemFieldJsonTag := valueItemField.Tag.Get("json"); len(valueItemFieldJsonTag) != 0 {
				valueKey := valueItemFieldJsonTag
				if strings.HasSuffix(valueItemFieldJsonTag, ",inline") {
					valueKey = ""
				} else if strings.HasSuffix(valueItemFieldJsonTag, ",omitempty") {
					splits := strings.Split(valueItemFieldJsonTag, ",")
					valueKey = splits[0]
				}

				ret = append(ret, bsonElement(valueItemField.Type.Kind(), valueKey, value.Field(i))...)
			}
		}

		if len(key) != 0 {
			return []*bson.Element{bson.EC.SubDocumentFromElements(key, ret...)}
		}

		return ret
	}

	return nil
}

func bsonValue(kind reflect.Kind, value reflect.Value) *bson.Value {
	switch kind {
	case reflect.Uint32, reflect.Int32:
		return bson.VC.Int32(int32(value.Int()))
	case reflect.Uint64, reflect.Int64:
		return bson.VC.Int64(value.Int())
	case reflect.String:
		return bson.VC.String(value.String())
	case reflect.Bool:
		return bson.VC.Boolean(value.Bool())
	case reflect.Float32, reflect.Float64:
		return bson.VC.Double(value.Float())
	case reflect.Map:
		if value.IsNil() {
			return nil
		}

		subRoot := newIgnoreNilInsertBsonDocument()

		for _, key := range value.MapKeys() {
			valueValue := value.MapIndex(key)
			subRoot.Append(bsonElement(valueValue.Kind(), key.String(), valueValue)...)
		}

		return bson.VC.Document(subRoot)
	case reflect.Slice:
		if value.IsNil() {
			return nil
		}

		subRoot := bson.ArrayFromDocument(newIgnoreNilInsertBsonDocument())
		for i := 0; i < value.Len(); i++ {
			valueValue := value.Index(i)
			subRoot.Append(
				bsonValue(valueValue.Kind(), valueValue),
			)
		}

		return bson.VC.Array(subRoot)
	case reflect.Interface:
		if value.IsNil() {
			return nil
		}

		return bsonValue(value.Kind(), value.Elem())
	case reflect.Ptr:
		value := value.Elem()

		if !value.IsValid() {
			return nil
		}

		return bsonValue(value.Kind(), value)
	case reflect.Struct:
		switch t := value.Interface().(type) {
		case apisMeta.Time:
			return bson.VC.DateTime(t.Time.Unix())
		case apisMeta.MicroTime:
			return bson.VC.DateTime(t.Time.Unix())
		case time.Time:
			return bson.VC.DateTime(t.Unix())
		case apisMeta.Timestamp:
			return bson.VC.DateTime(time.Unix(t.Seconds, int64(t.Nanos)).Unix())
		}

		ret := make([]*bson.Element, 0, 16)
		valueType := value.Type()

		for i := 0; i < value.NumField(); i++ {
			valueItemField := valueType.Field(i)
			if len(valueItemField.PkgPath) != 0 {
				continue
			}

			if valueItemFieldJsonTag := valueItemField.Tag.Get("json"); len(valueItemFieldJsonTag) != 0 {
				valueKey := valueItemFieldJsonTag
				if strings.HasSuffix(valueItemFieldJsonTag, ",inline") {
					valueKey = ""
				} else if strings.HasSuffix(valueItemFieldJsonTag, ",omitempty") {
					splits := strings.Split(valueItemFieldJsonTag, ",")
					valueKey = splits[0]
				}

				ret = append(ret, bsonElement(valueItemField.Type.Kind(), valueKey, value.Field(i))...)
			}
		}

		doc := newIgnoreNilInsertBsonDocument()
		doc.Append(ret...)

		return bson.VC.Document(doc)
	}

	return nil
}
