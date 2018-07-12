package pipes

import (
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/thxcode/kubernetes-event-exporter/pkg/simplelogger"
	apiCoreV1 "k8s.io/api/core/v1"
)

type loggerPipe struct {
	logContext logrus.Fields

	sync.Once
}

func (p *loggerPipe) Start() error {
	p.Do(func() {
		logrus.WithFields(p.logContext).Debugln("starting")
	})

	return nil
}

func (p *loggerPipe) Stop() {
	logrus.WithFields(p.logContext).Debugln("stopped")
}

func (p *loggerPipe) OnAdd(event *apiCoreV1.Event) error {
	logrus.WithFields(p.logContext).WithField("operation", "OnAdd").Debugln(showHead(printEvent(inState, event)))

	return nil
}

func (p *loggerPipe) OnUpdate(oldEvent *apiCoreV1.Event, newEvent *apiCoreV1.Event) error {
	logrus.WithFields(p.logContext).WithField("operation", "OnUpdate").Debugln(showHead(printEvent(outState, oldEvent), printEvent(inState, newEvent)))

	return nil
}

func (p *loggerPipe) OnDelete(event *apiCoreV1.Event) error {
	logrus.WithFields(p.logContext).WithField("operation", "OnDelete").Debugln(showHead(printEvent(outState, event)))

	return nil
}

func (p *loggerPipe) OnList(eventList *apiCoreV1.EventList) error {
	if len(eventList.Items) != 0 {
		logrus.WithFields(p.logContext).WithField("operation", "OnList").Debugln(showHead(printEventList(eventList)))
	}

	return nil
}

func NewLogger(khost string) *loggerPipe {
	return &loggerPipe{
		logContext: simplelogger.CreateLogContext("PIPE<logger>", khost),
	}
}
