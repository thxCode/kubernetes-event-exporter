package main

import (
	"sort"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thxcode/kubernetes-event-exporter/pkg/events"
	"github.com/thxcode/kubernetes-event-exporter/pkg/events/sinks"
	"github.com/thxcode/kubernetes-event-exporter/pkg/events/sinks/pipes"
	"github.com/thxcode/kubernetes-event-exporter/pkg/utils/logger"
	"github.com/thxcode/kubernetes-event-exporter/pkg/watchers"
	"k8s.io/client-go/kubernetes"
)

type eventExporter struct {
	logContext logrus.Fields

	watcher watchers.Watcher
	sink    sinks.Sink
}

func (e *eventExporter) Run(stopCh <-chan struct{}) {
	if err := e.sink.Run(stopCh); err != nil {
		logrus.WithFields(e.logContext).WithError(err).Fatalln("fail to run sink")
	}

	logrus.WithFields(e.logContext).Debugln("starting")
	e.watcher.Run(stopCh)
	logrus.WithFields(e.logContext).Debugln("stopped")
}

func newEventExporter(kclient kubernetes.Interface, khost string, resyncPeriod time.Duration, storageTTL time.Duration, usePipes []string, pipesParallel bool) *eventExporter {
	if len(usePipes) == 0 {
		logrus.Fatalln("failed to create sink, there aren't any pipes enabled")
	}

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
			ps = append(ps, pipes.NewMongoDB(khost, kclient))
		}
	}

	sink, err := sinks.NewDefaultSink(&sinks.DefaultSinkConfig{
		KubernetesHost: khost,
		Pipes:          ps,
		PipesParallel:  pipesParallel,
	})
	if err != nil {
		logrus.WithError(err).Fatalf("failed to create sink")
	}

	return &eventExporter{
		logContext: logger.CreateLogContext("EXPORTER", khost),
		watcher:    createWatcher(kclient, sink, resyncPeriod, storageTTL),
		sink:       sink,
	}
}

func createWatcher(client kubernetes.Interface, sink sinks.Sink, resyncPeriod time.Duration, storageTTL time.Duration) watchers.Watcher {
	return events.NewEventWatcher(client, &events.EventWatcherConfig{
		OnList:       sink.OnList,
		ResyncPeriod: resyncPeriod,
		StorageTTL:   storageTTL,
		Handler:      sink,
	})
}
