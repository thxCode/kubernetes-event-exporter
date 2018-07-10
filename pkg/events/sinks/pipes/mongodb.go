package pipes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/juju/errors"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/core/option"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/sirupsen/logrus"
	"github.com/thxcode/kubernetes-event-exporter/pkg/events/sinks"
	"github.com/thxcode/kubernetes-event-exporter/pkg/utils/logger"
	apiCoreV1 "k8s.io/api/core/v1"
	apisMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	MongodbConnectURIEnvKey       = "PIPE_MONGODB_CONNECT_URI"
	MongodbDatabaseNameEnvKey     = "PIPE_MONGODB_DATABASE_NAME"
	MongodbEnableJsonAttachEnvKey = "PIPE_MONGODB_ENABLE_JSON_ATTACH"

	dataOpenIdKey     = "_id"
	dataAttachJsonKey = "_attachJson"
	dataAttachDocKey  = "_attachDoc"
)

type eventChanUnit struct {
	eventBson   *bson.Document
	eventHandle sinks.Handle
}

type mongodbPipe struct {
	logContext logrus.Fields

	rootCtx        context.Context
	rootCancelFunc context.CancelFunc
	eventChan      chan eventChanUnit
	eventChanStop  chan chan struct{}
	kclient        kubernetes.Interface

	mongoCollection  *mongo.Collection
	mongoDatabase    *mongo.Database
	mongoClient      *mongo.Client
	enableJsonAttach bool

	sync.RWMutex
	sync.Once
}

func (p *mongodbPipe) Start() (err error) {
	p.Do(func() {
		logrus.WithFields(p.logContext).Debugln("starting")

		uri := os.Getenv(MongodbConnectURIEnvKey)
		if len(uri) == 0 {
			err = errors.Errorf(`"%s" env is required`, MongodbConnectURIEnvKey)
			return
		} else {
			enableJsonAttachEnv := os.Getenv(MongodbEnableJsonAttachEnvKey)
			p.enableJsonAttach = strings.ToLower(enableJsonAttachEnv) == "true"
			if p.enableJsonAttach {
				logrus.WithFields(p.logContext).Debugln("enabling Pod or Node info json form attaching")
			}

			p.mongoClient, err = mongo.Connect(p.rootCtx, uri, nil)
			if err != nil {
				err = errors.Annotate(err, "MongoDB fail to create client")
				return
			}

			dbname := os.Getenv(MongodbDatabaseNameEnvKey)
			if len(dbname) == 0 {
				dbname = "kubernetes_events"
			}
			p.mongoDatabase = p.mongoClient.Database(dbname)
			logrus.WithFields(p.logContext).Debugf("using %s database", dbname)

			khost := (p.logContext[logger.LogKubernetesHostKey]).(string)
			if len(khost) == 0 {
				err = errors.Errorf("can't get Kubernetes host")
				return
			}

			// get mongodb collection name
			collectionsMapCollection := p.mongoDatabase.Collection("collections_map")
			collectionsMapCollection.Indexes().CreateMany(p.rootCtx, nil,
				mongo.IndexModel{
					Keys: bson.NewDocument(
						bson.EC.Int64("kubernetes_host", 1),
					),
					Options: bson.NewDocument(
						bson.EC.Boolean("background", true),
						bson.EC.Boolean("unique", true),
						bson.EC.String("name", "khost"),
					),
				},
				mongo.IndexModel{
					Keys: bson.NewDocument(
						bson.EC.Int64("collection_name", 1),
					),
					Options: bson.NewDocument(
						bson.EC.Boolean("background", true),
						bson.EC.Boolean("unique", true),
						bson.EC.String("name", "colname"),
					),
				},
			)

			docResult := collectionsMapCollection.FindOne(
				p.rootCtx,
				bson.NewDocument(
					bson.EC.String("kubernetes_host", khost),
				),
			)
			colname := ""
			storageCollectionMap := bson.NewDocument()

			if err := docResult.Decode(storageCollectionMap); err != nil {
				if err != mongo.ErrNoDocuments {
					err = errors.Annotatef(err, "can't find info from %s.collections_map collection", dbname)
					return
				} else {
					colname = hashing([]byte(khost))[:16]
					if _, err = collectionsMapCollection.InsertOne(
						p.rootCtx,
						bson.NewDocument(
							bson.EC.String("kubernetes_host", khost),
							bson.EC.String("collection_name", colname),
						),
					); err != nil {
						err = errors.Annotatef(err, "can't insert info into %s.collections_map collection", dbname)
						return
					}
				}
			} else {
				if colnameVal := storageCollectionMap.Lookup("collection_name"); colnameVal == nil {
					err = errors.Annotatef(err, "can't find collection_name column on %s.collections_map collection schema", dbname)
					return
				} else {
					colname = colnameVal.StringValue()
				}
			}

			if len(colname) == 0 {
				err = errors.New(fmt.Sprintf("can't use blank collection name for %s", khost))
				return
			}

			p.mongoCollection = p.mongoDatabase.Collection(colname)
			p.mongoCollection.Indexes().CreateMany(p.rootCtx, nil,
				mongo.IndexModel{
					Keys: bson.NewDocument(
						bson.EC.Int64("metadata.uid", 1),
					),
					Options: bson.NewDocument(
						bson.EC.Boolean("background", true),
						bson.EC.Boolean("unique", true),
						bson.EC.String("name", "query_id"),
					),
				},
				mongo.IndexModel{
					Keys: bson.NewDocument(
						bson.EC.Int64("involvedObject.kind", 1),
						bson.EC.Int64("involvedObject.name", 1),
						bson.EC.Int64("involvedObject.namespace", 1),
					),
					Options: bson.NewDocument(
						bson.EC.Boolean("background", true),
						bson.EC.Boolean("unique", false),
						bson.EC.Boolean("sparse", true),
						bson.EC.String("name", "query_info"),
					),
				},
				mongo.IndexModel{
					Keys: bson.NewDocument(
						bson.EC.Int64("metadata.creationTimestamp", -1),
					),
					Options: bson.NewDocument(
						bson.EC.Boolean("background", true),
						bson.EC.Boolean("unique", false),
						bson.EC.Boolean("sparse", true),
						bson.EC.String("name", "query_time"),
					),
				},
			)

			go p.dealEventChan()

		}
	})

	return err
}

func (p *mongodbPipe) Stop() {
	logrus.WithFields(p.logContext).Debugln("stopping")

	<-p.flushEventChan()
	p.mongoClient.Disconnect(p.rootCtx)
	p.mongoDatabase = nil
	p.mongoCollection = nil
	p.rootCancelFunc()

	logrus.WithFields(p.logContext).Debugln("stopped")
}

func (p *mongodbPipe) OnAdd(event *apiCoreV1.Event) error {
	p.RLock()
	defer p.RUnlock()

	involvedObject := event.InvolvedObject
	kind := involvedObject.Kind
	namespace := involvedObject.Namespace
	name := involvedObject.Name

	var bufferEventBson *bson.Document
	switch kind {
	case "Pod":
		bufferEventBson = eventToBson(event)

		// scrape Pod info
		podInfo, err := p.kclient.CoreV1().Pods(namespace).Get(name, apisMetaV1.GetOptions{})
		if err != nil {
			return err
		}
		podInfoJson, err := json.Marshal(podInfo)
		if err != nil {
			return err
		}

		if p.enableJsonAttach {
			bufferEventBson.Append(
				bson.EC.String(dataAttachJsonKey, *(*string)(unsafe.Pointer(&podInfoJson))),
			)
		} else {
			podInfoBson, err := bson.ParseExtJSONObject(*(*string)(unsafe.Pointer(&podInfoJson)))
			if err != nil {
				return err
			}

			bufferEventBson.Append(
				bson.EC.SubDocument(dataAttachDocKey, podInfoBson),
			)
		}
	case "Node":
		bufferEventBson = eventToBson(event)

		// scrape Node info
		nodeInfo, err := p.kclient.CoreV1().Nodes().Get(name, apisMetaV1.GetOptions{})
		if err != nil {
			return err
		}
		nodeInfoJson, err := json.Marshal(nodeInfo)
		if err != nil {
			return err
		}

		if p.enableJsonAttach {
			bufferEventBson.Append(
				bson.EC.String(dataAttachJsonKey, *(*string)(unsafe.Pointer(&nodeInfoJson))),
			)
		} else {
			nodeInfoBson, err := bson.ParseExtJSONObject(*(*string)(unsafe.Pointer(&nodeInfoJson)))
			if err != nil {
				return err
			}

			bufferEventBson.Append(
				bson.EC.SubDocument(dataAttachDocKey, nodeInfoBson),
			)
		}
	default:
		logrus.WithFields(p.logContext).Debugf("ignoring the addition operation for %s", kind)
	}

	if bufferEventBson != nil {
		p.eventChan <- eventChanUnit{
			bufferEventBson,
			sinks.OnAdd,
		}
	}

	return nil
}

func (p *mongodbPipe) OnUpdate(_ *apiCoreV1.Event, event *apiCoreV1.Event) error {
	p.RLock()
	defer p.RUnlock()

	involvedObject := event.InvolvedObject
	kind := involvedObject.Kind

	var bufferEventBson *bson.Document
	switch kind {
	case "Pod", "Node":
		bufferEventBson = eventToBson(event)
	default:
		logrus.WithFields(p.logContext).Debugf("ignoring the updating operation for %s", kind)
	}

	if bufferEventBson != nil {
		p.eventChan <- eventChanUnit{
			bufferEventBson,
			sinks.OnUpdate,
		}
	}

	return nil

}

func (p *mongodbPipe) OnDelete(event *apiCoreV1.Event) error {
	logrus.WithFields(p.logContext).Debugln("ignoring the deletion operation")
	return nil
}

func (p *mongodbPipe) OnList(eventList *apiCoreV1.EventList) error {
	p.RLock()
	defer p.RUnlock()

	for _, event := range eventList.Items {
		involvedObject := event.InvolvedObject
		kind := involvedObject.Kind

		var bufferEventBson *bson.Document
		switch kind {
		case "Pod", "Node":
			bufferEventBson = eventToBson(&event)
		default:
			logrus.WithFields(p.logContext).Debugf("ignoring the listing operation for %s", kind)
		}

		if bufferEventBson != nil {
			p.eventChan <- eventChanUnit{
				bufferEventBson,
				sinks.OnList,
			}
		}
	}

	return nil
}

func (p *mongodbPipe) flushEventChan() <-chan struct{} {
	ch := make(chan struct{})
	p.eventChanStop <- ch
	return ch
}

func (p *mongodbPipe) dealEventChan() {
loop:
	for {
		select {
		case unit := <-p.eventChan:
			p.dealingEvent(&unit)

		case done := <-p.eventChanStop:
			// get the lock and prevent new event
			p.Lock()

			// deal all caching event
			for unit := range p.eventChan {
				p.dealingEvent(&unit)
			}

			close(p.eventChan)
			close(done)
			p.Unlock()

			break loop
		}
	}
}

func (p *mongodbPipe) dealingEvent(unit *eventChanUnit) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				logrus.WithFields(p.logContext).WithError(err).Errorln("failed to deal event")
			} else {
				logrus.WithFields(p.logContext).Errorln("failed to deal event,", r)
			}
		}
	}()

	if unit == nil {
		return
	}

	eventBson := unit.eventBson
	metadataUidElement, err := eventBson.LookupElementErr("metadata", "uid")
	if err != nil {
		panic(errors.New(`the "metadata.uid" is required`))
	}
	metadataUid := metadataUidElement.Value().StringValue()

	switch unit.eventHandle {
	case sinks.OnList:
		ret := p.mongoCollection.FindOne(
			p.rootCtx,
			bson.NewDocument(
				bson.EC.String("metadata.uid", metadataUid),
			),
			option.OptProjection{
				Projection: bson.NewDocument(
					bson.EC.Boolean(dataOpenIdKey, false),
					bson.EC.Boolean(dataAttachJsonKey, false),
					bson.EC.Boolean(dataAttachDocKey, false),
				),
			},
		)
		inDoc := bson.NewDocument()
		if err := ret.Decode(inDoc); err != nil {
			if err != mongo.ErrNoDocuments {
				panic(errors.Annotatef(err, "can't find \n%s", eventBson.ToExtJSON(true)))
			} else {
				_, err := p.mongoCollection.InsertOne(
					p.rootCtx,
					eventBson,
				)
				if err != nil {
					panic(errors.Annotatef(err, "can't insert \n%s", eventBson.ToExtJSON(true)))
				} else {
					logrus.WithFields(p.logContext).Debugln("success add event:", metadataUid)
				}
			}
		} else {
			if !inDoc.Equal(eventBson) {
				_, err := p.mongoCollection.UpdateOne(
					p.rootCtx,
					bson.NewDocument(
						bson.EC.String("metadata.uid", metadataUid),
					),
					eventBson,
				)
				if err != nil {
					panic(errors.Annotatef(err, "can't update \n%s", eventBson.ToExtJSON(true)))
				} else {
					logrus.WithFields(p.logContext).Debugln("success update event:", metadataUid)
				}
			}
		}
	case sinks.OnAdd:
		_, err := p.mongoCollection.InsertOne(
			p.rootCtx,
			eventBson,
		)
		if err != nil {
			panic(errors.Annotatef(err, "can't insert \n%s", eventBson.ToExtJSON(true)))
		} else {
			logrus.WithFields(p.logContext).Debugln("success add event:", metadataUid)
		}
	case sinks.OnUpdate:
		ret := p.mongoCollection.FindOneAndUpdate(
			p.rootCtx,
			bson.NewDocument(
				bson.EC.String("metadata.uid", metadataUid),
			),
			eventBson,
			option.OptProjection{
				Projection: bson.NewDocument(
					bson.EC.Boolean("_id", true),
				),
			},
			option.OptMaxTime(10*time.Second),
		)
		if err := ret.Decode(nil); err != nil {
			panic(errors.Annotatef(err, "can't update \n%s", eventBson.ToExtJSON(true)))
		} else {
			logrus.WithFields(p.logContext).Debugln("success update event:", metadataUid)
		}
	}
}

func NewMongoDB(khost string, kclient kubernetes.Interface) *mongodbPipe {
	ctx, cancelFunc := context.WithCancel(context.Background())

	return &mongodbPipe{
		logContext: logger.CreateLogContext("PIPE<mongodb>", khost),

		rootCtx:        ctx,
		rootCancelFunc: cancelFunc,
		eventChan:      make(chan eventChanUnit, 1<<20),
		eventChanStop:  make(chan chan struct{}),

		kclient: kclient,
	}
}

func hashing(bytes []byte) string {
	hasher := sha256.New()
	hasher.Write(bytes)

	return hex.EncodeToString(hasher.Sum(nil))
}
