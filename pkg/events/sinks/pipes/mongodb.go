package pipes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/juju/errors"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/core/option"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/sirupsen/logrus"
	"github.com/thxcode/kubernetes-event-exporter/pkg/events/sinks"
	"github.com/thxcode/kubernetes-event-exporter/pkg/simplelogger"
	apiCoreV1 "k8s.io/api/core/v1"
	apisMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	MongodbConnectURIEnvKey   = "PIPE_MONGODB_CONNECT_URI"
	MongodbDatabaseNameEnvKey = "PIPE_MONGODB_DATABASE_NAME"

	dataOpenIdKey     = "_id"
	dataAttachPodKey  = "attachPod"
	dataAttachNodeKey = "attachNode"
)

type eventChanUnit struct {
	eventBson   *bson.Document
	eventHandle sinks.Handle
}

type mongodbPipe struct {
	logContext logrus.Fields

	ctx           context.Context
	cancelFunc    context.CancelFunc
	eventChan     chan eventChanUnit
	eventChanStop chan chan struct{}
	kclient       kubernetes.Interface

	mongoCollection *mongo.Collection
	mongoDatabase   *mongo.Database
	mongoClient     *mongo.Client

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
			khost := (p.logContext[simplelogger.LogKubernetesHostKey]).(string)
			if len(khost) == 0 {
				err = errors.Errorf("can't get Kubernetes host")
				return
			}

			p.mongoClient, err = mongo.Connect(p.ctx, uri, nil)
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

			clusterID := khost[(strings.LastIndex(khost, "/") + 1):]
			logrus.WithFields(p.logContext).Debugf("using %s collection", clusterID)

			p.mongoCollection = p.mongoDatabase.Collection(clusterID)
			p.mongoCollection.Indexes().CreateMany(p.ctx, nil,
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
	p.mongoClient.Disconnect(p.ctx)
	p.mongoDatabase = nil
	p.mongoCollection = nil
	p.cancelFunc()

	logrus.WithFields(p.logContext).Debugln("stopped")
}

func (p *mongodbPipe) doAttachForEvent(kind, namespace, name string, eventLog *HuaWeiEventLogBson) error {
	switch kind {
	case "Pod":
		// scrape Pod info
		podInfo, err := p.kclient.CoreV1().Pods(namespace).Get(name, apisMetaV1.GetOptions{})
		if err != nil {
			return err
		}

		eventLog.AttachPod = podInfo
	case "Node":
		// scrape Node info
		nodeInfo, err := p.kclient.CoreV1().Nodes().Get(name, apisMetaV1.GetOptions{})
		if err != nil {
			return err
		}

		eventLog.AttachNode = nodeInfo
	default:
		logrus.WithFields(p.logContext).Debugf("ignoring the addition operation for %s", kind)
	}

	return nil
}

func (p *mongodbPipe) OnAdd(event *apiCoreV1.Event) error {
	p.RLock()
	defer p.RUnlock()

	involvedObject := event.InvolvedObject
	kind := involvedObject.Kind
	namespace := involvedObject.Namespace
	name := involvedObject.Name

	var wrapEvent *HuaWeiEventLogBson
	switch kind {
	case "Pod":
		wrapEvent = &HuaWeiEventLogBson{
			Event: event,
		}
	case "Node":
		wrapEvent = &HuaWeiEventLogBson{
			Event: event,
		}
	default:
		logrus.WithFields(p.logContext).Debugf("ignoring the addition operation for %s", kind)
	}

	if wrapEvent != nil {
		if err := p.doAttachForEvent(kind, namespace, name, wrapEvent); err != nil {
			return err
		}

		p.eventChan <- eventChanUnit{
			wrapEvent.MustMarshalBSONDocument(),
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

	var wrapEvent *HuaWeiEventLogBson
	switch kind {
	case "Pod", "Node":
		wrapEvent = &HuaWeiEventLogBson{
			Event: event,
		}
	default:
		logrus.WithFields(p.logContext).Debugf("ignoring the updating operation for %s", kind)
	}

	if wrapEvent != nil {
		p.eventChan <- eventChanUnit{
			wrapEvent.MustMarshalBSONDocument(),
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

		var wrapEvent *HuaWeiEventLogBson
		switch kind {
		case "Pod", "Node":
			wrapEvent = &HuaWeiEventLogBson{
				Event: &event,
			}
		default:
			logrus.WithFields(p.logContext).Debugf("ignoring the listing operation for %s", kind)
		}

		if wrapEvent != nil {
			p.eventChan <- eventChanUnit{
				wrapEvent.MustMarshalBSONDocument(),
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

	metadataUidElement := eventBson.LookupElement("metadata", "uid")
	metadataUid := metadataUidElement.Value().StringValue()

	switch unit.eventHandle {
	case sinks.OnList:
		ret := p.mongoCollection.FindOne(
			p.ctx,
			bson.NewDocument(
				bson.EC.String("metadata.uid", metadataUid),
			),
			option.OptProjection{
				Projection: bson.NewDocument(
					bson.EC.Boolean(dataOpenIdKey, false),
					bson.EC.Boolean(dataAttachPodKey, false),
					bson.EC.Boolean(dataAttachNodeKey, false),
				),
			},
		)
		inDoc := bson.NewDocument()
		if err := ret.Decode(inDoc); err != nil {
			if err != mongo.ErrNoDocuments {
				panic(errors.Annotatef(err, "can't find \n%s", eventBson.ToExtJSON(true)))
			} else {
				// // onAdd
				// involvedObjectKindElement := eventBson.LookupElement("involvedObject", "kind")
				// involvedObjectKind := involvedObjectKindElement.Value().StringValue()
				// involvedObjectNamespaceElement := eventBson.LookupElement("involvedObject", "namespace")
				// involvedObjectNamespace := involvedObjectNamespaceElement.Value().StringValue()
				// involvedObjectNameElement := eventBson.LookupElement("involvedObject", "name")
				// involvedObjectName := involvedObjectNameElement.Value().StringValue()
				//
				// if err := p.doAttachForEvent(involvedObjectKind, involvedObjectNamespace, involvedObjectName, eventBson); err != nil {
				// 	panic(errors.Annotatef(err, "can't attach info for \n%s", eventBson.ToExtJSON(true)))
				// }
				//
				// if _, err := p.mongoCollection.InsertOne(
				// 	p.ctx,
				// 	eventBson,
				// ); err != nil {
				// 	panic(errors.Annotatef(err, "can't insert \n%s", eventBson.ToExtJSON(true)))
				// } else {
				// 	logrus.WithFields(p.logContext).Debugln("success add event:", metadataUid)
				// }
			}
		} else {
			if !inDoc.Equal(eventBson) {
				_, err := p.mongoCollection.UpdateOne(
					p.ctx,
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
			p.ctx,
			eventBson,
		)
		if err != nil {
			panic(errors.Annotatef(err, "can't insert \n%s", eventBson.ToExtJSON(true)))
		} else {
			logrus.WithFields(p.logContext).Debugln("success add event:", metadataUid)
		}
	case sinks.OnUpdate:
		ret := p.mongoCollection.FindOneAndUpdate(
			p.ctx,
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

func NewMongoDB(rootCtx context.Context, khost string, kclient kubernetes.Interface) *mongodbPipe {
	ctx, cancelFunc := context.WithCancel(rootCtx)

	return &mongodbPipe{
		logContext: simplelogger.CreateLogContext("PIPE<mongodb>", khost),

		ctx:           ctx,
		cancelFunc:    cancelFunc,
		eventChan:     make(chan eventChanUnit, 1<<20),
		eventChanStop: make(chan chan struct{}),

		kclient: kclient,
	}
}

func hashing(bytes []byte) string {
	hasher := sha256.New()
	hasher.Write(bytes)

	return hex.EncodeToString(hasher.Sum(nil))
}
