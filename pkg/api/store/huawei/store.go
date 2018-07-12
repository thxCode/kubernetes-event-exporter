package huawei

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/juju/errors"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/core/option"
	"github.com/mongodb/mongo-go-driver/mongo"
	normanHttpError "github.com/rancher/norman/httperror"
	normanTypes "github.com/rancher/norman/types"
	exporterEventPipes "github.com/thxcode/kubernetes-event-exporter/pkg/events/sinks/pipes"
)

func CreateMongo(ctx context.Context, wg *sync.WaitGroup, schema *normanTypes.Schema) {
	wg.Add(1)

	go func() {
		defer wg.Done()

	}()

	uri := os.Getenv(exporterEventPipes.MongodbConnectURIEnvKey)
	if len(uri) == 0 {
		panic(errors.Errorf(`"%s" env is required`, exporterEventPipes.MongodbConnectURIEnvKey))
	} else {
		mongoClient, err := mongo.Connect(ctx, uri, nil)
		if err != nil {
			errors.Annotate(err, "MongoDB fail to create client")
		}

		dbname := os.Getenv(exporterEventPipes.MongodbDatabaseNameEnvKey)
		if len(dbname) == 0 {
			dbname = "kubernetes_events"
		}
		mongoDatabase := mongoClient.Database(dbname)

		schema.Store = &mongoStore{
			ctx,
			mongoClient,
			mongoDatabase,
		}
	}
}

type mongoStore struct {
	ctx           context.Context
	mongoClient   *mongo.Client
	mongoDatabase *mongo.Database
}

func (s *mongoStore) Context() normanTypes.StorageContext {
	return "mongo"
}

func (s *mongoStore) ByID(apiContext *normanTypes.APIContext, schema *normanTypes.Schema, id string) (map[string]interface{}, error) {
	return nil, normanHttpError.NewAPIError(normanHttpError.InvalidAction, "Unsupported action")
}

func (s *mongoStore) Create(apiContext *normanTypes.APIContext, schema *normanTypes.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	return nil, normanHttpError.NewAPIError(normanHttpError.InvalidAction, "Unsupported action")
}

func (s *mongoStore) Update(apiContext *normanTypes.APIContext, schema *normanTypes.Schema, data map[string]interface{}, id string) (map[string]interface{}, error) {
	return nil, normanHttpError.NewAPIError(normanHttpError.InvalidAction, "Unsupported action")
}

func (s *mongoStore) Delete(apiContext *normanTypes.APIContext, schema *normanTypes.Schema, id string) (map[string]interface{}, error) {
	return nil, normanHttpError.NewAPIError(normanHttpError.InvalidAction, "Unsupported action")
}

func (s *mongoStore) Watch(apiContext *normanTypes.APIContext, schema *normanTypes.Schema, opt *normanTypes.QueryOptions) (chan map[string]interface{}, error) {
	return nil, normanHttpError.NewAPIError(normanHttpError.InvalidAction, "Unsupported action")
}

func (s *mongoStore) List(apiContext *normanTypes.APIContext, schema *normanTypes.Schema, opt *normanTypes.QueryOptions) ([]map[string]interface{}, error) {
	var (
		eventId               string
		resourceNamespace     string
		resourceKind          string
		logType               string
		createdRangeFormat    string
		createdRangeStart     string
		createdRangeEnd       string
		createdRangeSortOrder string

		mongoFilter = bson.NewDocument()
		mongoLimit  option.OptLimit
		mongoSort   option.OptSort
	)

	// take conditions
	for _, condition := range opt.Conditions {
		switch condition.Field {
		case "eventId":
			eventId = condition.Value
		case "namespaceId":
			resourceNamespace = condition.Value
		case "resourceKind":
			resourceKind = condition.Value
		case "logType":
			logType = condition.Value
		case "createdRangeFormat":
			createdRangeFormat = condition.Value
		case "createdRangeStart":
			createdRangeStart = condition.Value
		case "createdRangeEnd":
			createdRangeEnd = condition.Value
		case "order":
			createdRangeSortOrder = condition.Value
		}
	}

	// check all
	if len(eventId) == 0 {
		return nil, nil
	}

	var (
		createdRangeEndTime   *time.Time
		createdRangeStartTime *time.Time
	)
	if len(createdRangeEnd) != 0 {
		ret, err := time.Parse(createdRangeFormat, createdRangeEnd)
		if err != nil {
			return nil, normanHttpError.WrapAPIError(err, normanHttpError.InvalidDateFormat, fmt.Sprintf("fail to parse createdRangeEnd by format: %s", createdRangeFormat))
		}

		createdRangeEndTime = &ret
	}
	if len(createdRangeStart) != 0 {
		ret, err := time.Parse(createdRangeFormat, createdRangeStart)
		if err != nil {
			return nil, normanHttpError.WrapAPIError(err, normanHttpError.InvalidDateFormat, fmt.Sprintf("fail to parse createdRangeStart by format: %s", createdRangeFormat))
		}

		createdRangeStartTime = &ret
		if createdRangeEndTime != nil && createdRangeStartTime.After(*createdRangeEndTime) {
			return nil, normanHttpError.NewAPIError(normanHttpError.InvalidOption, "createRangeEnd is before createRangeStart")
		}
	}
	if createdRangeStartTime != nil {
		if createdRangeEndTime != nil {
			mongoFilter.Append(
				bson.EC.SubDocumentFromElements("metadata.creationTimestamp",
					bson.EC.Time("$ge", *createdRangeStartTime),
					bson.EC.Time("$le", *createdRangeEndTime),
				),
			)
		} else {
			mongoFilter.Append(
				bson.EC.SubDocumentFromElements("metadata.creationTimestamp",
					bson.EC.Time("$ge", *createdRangeStartTime),
				),
			)
		}
	} else if createdRangeEndTime != nil {
		mongoFilter.Append(
			bson.EC.SubDocumentFromElements("metadata.creationTimestamp",
				bson.EC.Time("$le", *createdRangeEndTime),
			),
		)
	}

	if len(resourceNamespace) != 0 {
		mongoFilter.Append(bson.EC.String("involvedObject.namespace", resourceNamespace))
	}

	if resourceKind == "Pod" || resourceKind == "Node" || resourceKind == "Container" {
		mongoFilter.Append(bson.EC.String("involvedObject.kind", resourceKind))
	}

	if logType == "Normal" || logType == "Warning" {
		mongoFilter.Append(bson.EC.String("type", logType))
	}

	if len(createdRangeFormat) == 0 {
		createdRangeFormat = "2006-01-02T15:04:05Z07:00"
	}

	if createdRangeSortOrder != "ASC" {
		mongoSort = option.OptSort{
			Sort: bson.NewDocument(
				bson.EC.Int32("metadata.creationTimestamp", 1),
			),
		}
	} else {
		mongoSort = option.OptSort{
			Sort: bson.NewDocument(
				bson.EC.Int32("metadata.creationTimestamp", -1),
			),
		}
	}

	// take pagination
	if opt.Pagination.Limit == nil {
		mongoLimit = option.OptLimit(1000)
	} else {
		mongoLimit = option.OptLimit(*opt.Pagination.Limit)
	}
	if len(opt.Pagination.Marker) != 0 {
		mongoFilter.Append(bson.EC.SubDocumentFromElements("_id",
			bson.EC.String("$gt", opt.Pagination.Marker),
		))
	}

	mongoCollection := s.mongoDatabase.Collection(eventId)
	ctx, _ := context.WithTimeout(s.ctx, 10*time.Second)

	docCursor, err := mongoCollection.Find(
		ctx,
		mongoFilter,
		mongoLimit,
		mongoSort,
	)
	if err != nil {
		return nil, normanHttpError.WrapAPIError(err, normanHttpError.ServerError, "can't find event logs")
	}

	ret := make([]map[string]interface{}, 0, mongoLimit)
	doc := bson.NewDocument()
	for docCursor.Next(ctx) {
		doc.Reset()

		err := docCursor.Decode(doc)
		if err != nil {
			continue
		}

		docJson := doc.ToExtJSON(true)
		var docJsonObj map[string]interface{}
		if err := json.Unmarshal([]byte(docJson), &docJsonObj); err != nil {
			ret = append(ret, docJsonObj)
		}
	}

	return ret, nil
}
