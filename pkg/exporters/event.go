package exporters

import (
	"context"
	"sort"
	"time"

	"github.com/juju/errors"
	"github.com/sirupsen/logrus"
	"github.com/thxcode/kubernetes-event-exporter/pkg/events"
	"github.com/thxcode/kubernetes-event-exporter/pkg/events/sinks"
	"github.com/thxcode/kubernetes-event-exporter/pkg/events/sinks/pipes"
	"github.com/thxcode/kubernetes-event-exporter/pkg/simplelogger"
	"github.com/thxcode/kubernetes-event-exporter/pkg/watchers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type EventExporterConfig struct {
	ResyncPeriod  time.Duration
	StorageTTL    time.Duration
	UsePipes      []string
	PipesParallel bool

	WatchingK8sConfig *rest.Config
}

type EventExporter struct {
	logContext logrus.Fields

	watcher watchers.Watcher
	sink    sinks.Sink

	subCtx         context.Context
	subCtxCancelFn context.CancelFunc
}

func (e *EventExporter) Start() error {
	stopCh := e.subCtx.Done()
	if err := e.sink.Run(stopCh); err != nil {
		return errors.Annotate(err, "fail to run sink")
	}

	logrus.WithFields(e.logContext).Debugln("starting")
	go e.watcher.Run(stopCh)

	return nil
}

func (e *EventExporter) Stop() {
	e.subCtxCancelFn()
	logrus.WithFields(e.logContext).Debugln("stopped")
}

func NewEventExporter(config EventExporterConfig) (*EventExporter, error) {
	kHost := config.WatchingK8sConfig.Host
	kclient, err := kubernetes.NewForConfig(config.WatchingK8sConfig)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to create Kubernetes client for %s", kHost)
	}

	return newEventExporter(
		kclient,
		kHost,
		config.ResyncPeriod,
		config.StorageTTL,
		config.UsePipes,
		config.PipesParallel,
	)
}

func newEventExporter(kclient kubernetes.Interface, khost string, resyncPeriod time.Duration, storageTTL time.Duration, usePipes []string, pipesParallel bool) (*EventExporter, error) {
	if len(usePipes) == 0 {
		return nil, errors.New("failed to create sink, there aren't any pipes enabled")
	}

	subCtx, subCtxCancelFn := context.WithCancel(context.Background())

	usePipeSet := make(map[string]struct{}, len(usePipes))
	for _, usePipe := range usePipes {
		if _, ok := usePipeSet[usePipe]; !ok {
			usePipeSet[usePipe] = struct{}{}
		}
	}

	distinctUsePipe := make([]string, 0, len(usePipeSet))
	for p := range usePipeSet {
		distinctUsePipe = append(distinctUsePipe, p)
	}
	sort.Strings(distinctUsePipe)

	ps := make([]sinks.Pipe, 0, len(distinctUsePipe))
	for _, usePipe := range distinctUsePipe {
		switch usePipe {
		case "logger":
			if logrus.GetLevel() == logrus.DebugLevel {
				ps = append(ps, pipes.NewLogger(khost))
			}
		case "mongodb":
			ps = append(ps, pipes.NewMongoDB(subCtx, khost, kclient))
		}
	}

	sink, err := sinks.NewDefaultSink(&sinks.DefaultSinkConfig{
		KubernetesHost: khost,
		Pipes:          ps,
		PipesParallel:  pipesParallel,
	})
	if err != nil {
		return nil, errors.New("failed to create default sink")
	}

	return &EventExporter{
		logContext:     simplelogger.CreateLogContext("EXPORTER", khost),
		watcher:        createWatcher(kclient, sink, resyncPeriod, storageTTL),
		sink:           sink,
		subCtx:         subCtx,
		subCtxCancelFn: subCtxCancelFn,
	}, nil
}

func createWatcher(client kubernetes.Interface, sink sinks.Sink, resyncPeriod time.Duration, storageTTL time.Duration) watchers.Watcher {
	return events.NewEventWatcher(client, &events.EventWatcherConfig{
		OnList:       sink.OnList,
		ResyncPeriod: resyncPeriod,
		StorageTTL:   storageTTL,
		Handler:      sink,
	})
}
