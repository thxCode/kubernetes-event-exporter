package sinks

import (
	"fmt"
	"time"

	"github.com/juju/errors"
	"github.com/sirupsen/logrus"
	"github.com/thxcode/kubernetes-event-exporter/pkg/events"
	"github.com/thxcode/kubernetes-event-exporter/pkg/utils/logger"
	apiCoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

type Handle uint64

const (
	OnAdd Handle = iota
	OnUpdate
	OnDelete
	OnList
)

// Sink interface represents a generic sink that is responsible for handling
// actions upon the event objects and filter the initial events list. Note,
// that OnAdd method from the EventHandler interface will only receive
// objects that were added during watching phase, not before. If sink wishes
// to process the latter additions, it should implement additional logic in
// the OnList method.
type Sink interface {
	events.EventHandler

	OnList(eventList *apiCoreV1.EventList)

	Run(stopCh <-chan struct{}) error
}

type DefaultSinkConfig struct {
	KubernetesHost string
	Pipes          []Pipe
	PipesParallel  bool
}

type DefaultSink struct {
	logContext logrus.Fields

	pipesMap        map[string]Pipe
	isPipesParallel bool
}

func (s *DefaultSink) OnAdd(event *apiCoreV1.Event) {
	g := wait.Group{}
	defer g.Wait()

	for pipeName, pipe := range s.pipesMap {
		if s.isPipesParallel {
			func(pipeName string, pipe Pipe) {
				g.Start(func() {
					if err := pipe.OnAdd(event); err != nil {
						logrus.WithFields(s.logContext).WithError(err).Errorf("%s error occur", pipeName)
					}
				})
			}(pipeName, pipe)
		} else {
			if err := pipe.OnAdd(event); err != nil {
				logrus.WithFields(s.logContext).WithError(err).Errorf("%s error occur", pipeName)
				break
			}
		}
	}
}

func (s *DefaultSink) OnUpdate(oldEvent *apiCoreV1.Event, newEvent *apiCoreV1.Event) {
	g := wait.Group{}
	defer g.Wait()

	for pipeName, pipe := range s.pipesMap {
		if s.isPipesParallel {
			func(pipeName string, pipe Pipe) {
				g.Start(func() {
					if err := pipe.OnUpdate(oldEvent, newEvent); err != nil {
						logrus.WithFields(s.logContext).WithError(err).Errorf("%s error occur", pipeName)
					}
				})
			}(pipeName, pipe)
		} else {
			if err := pipe.OnUpdate(oldEvent, newEvent); err != nil {
				logrus.WithFields(s.logContext).WithError(err).Errorf("%s error occur", pipeName)
				break
			}
		}
	}
}

func (s *DefaultSink) OnDelete(event *apiCoreV1.Event) {
	g := wait.Group{}
	defer g.Wait()

	for pipeName, pipe := range s.pipesMap {
		if s.isPipesParallel {
			func(pipeName string, pipe Pipe) {
				g.Start(func() {
					if err := pipe.OnDelete(event); err != nil {
						logrus.WithFields(s.logContext).WithError(err).Errorf("%s error occur", pipeName)
					}
				})
			}(pipeName, pipe)
		} else {
			if err := pipe.OnDelete(event); err != nil {
				logrus.WithFields(s.logContext).WithError(err).Errorf("%s error occur", pipeName)
				break
			}
		}
	}
}

func (s *DefaultSink) OnList(eventList *apiCoreV1.EventList) {
	g := wait.Group{}
	defer g.Wait()

	for pipeName, pipe := range s.pipesMap {
		if s.isPipesParallel {
			func(pipeName string, pipe Pipe) {
				g.Start(func() {
					if err := pipe.OnList(eventList); err != nil {
						logrus.WithFields(s.logContext).WithError(err).Errorf("%s error occur", pipeName)
					}
				})
			}(pipeName, pipe)
		} else {
			if err := pipe.OnList(eventList); err != nil {
				logrus.WithFields(s.logContext).WithError(err).Errorf("%s error occur", pipeName)
				break
			}
		}
	}
}

func (s *DefaultSink) Run(stopCh <-chan struct{}) error {
	for {
		select {
		case <-time.Tick(30 * time.Second):
			return errors.New("timeout on pipes starting")
		default:
			logrus.WithFields(s.logContext).Debugf("prepare pipes")
			for _, pipe := range s.pipesMap {
				if err := pipe.Start(); err != nil {
					return errors.Annotatef(err, "%T starting error", pipe)
				}
			}
			logrus.WithFields(s.logContext).Debugf("running pipes")

			go func() {
				<-stopCh

				logrus.WithFields(s.logContext).Debugf("stopping pipes")
				for _, pipe := range s.pipesMap {
					pipe.Stop()
				}
				logrus.WithFields(s.logContext).Debugf("stopped pipes")
			}()

			return nil
		}
	}
}

func NewDefaultSink(config *DefaultSinkConfig) (*DefaultSink, error) {
	pipesMap := make(map[string]Pipe, len(config.Pipes))

	for _, pipe := range config.Pipes {
		pipesMap[fmt.Sprintf("%T", pipe)] = pipe
	}

	return &DefaultSink{
		logContext:      logger.CreateLogContext("SINK", config.KubernetesHost),
		pipesMap:        pipesMap,
		isPipesParallel: config.PipesParallel,
	}, nil
}
