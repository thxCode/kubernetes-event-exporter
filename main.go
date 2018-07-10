package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thxcode/kubernetes-event-exporter/pkg/events/sinks/pipes"
	"github.com/thxcode/kubernetes-event-exporter/pkg/utils/logger"
	"github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/prometheus/common/version"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func newSystemStopChannel() chan struct{} {
	ch := make(chan struct{})
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, os.Kill)
		sig := <-c
		logrus.Debugf("recieved signal %s, terminating", sig.String())

		close(ch)
	}()

	return ch
}

func initLog(c *cli.Context) {
	switch strings.ToLower(c.String("log-level")) {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	case "panic":
		logrus.SetLevel(logrus.PanicLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	switch c.String("log-format") {
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{})
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(logger.NewSimpleFormatter())
	}

	logrus.SetOutput(os.Stdout)
}

func main() {
	app := cli.NewApp()
	app.Name = "kubernetes-event-exporter"
	app.Version = version.Print("kubernetes-event-exporter")
	app.Usage = "An exporter exposes events of Kubernetes."
	app.Action = appAction

	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:   "kubeconfig",
			Usage:  "kube config for accessing Kubernetes cluster",
			EnvVar: "KUBECONFIG",
			Value:  &cli.StringSlice{},
		},
		cli.StringFlag{
			Name:   "log-level",
			Usage:  "log level for logurs",
			EnvVar: "LOG_LEVEL",
			Value:  "debug",
		},
		cli.StringFlag{
			Name:   "log-format",
			Usage:  "log formatter used (json, text, simple) for logrus",
			EnvVar: "LOG_FORMAT",
			Value:  "simple",
		},
		cli.DurationFlag{
			Name:   "resync-period",
			Usage:  "period for resynchronization on Kubernetes client-go cache reflector",
			EnvVar: "RESYNC_PERIOD",
			Value:  1 * time.Minute,
		},
		cli.DurationFlag{
			Name:   "storage-ttl",
			Usage:  "TTL for storage on Kubernetes client-go cache reflector",
			EnvVar: "STORAGE_TTL",
			Value:  2 * time.Hour,
		},
		cli.StringSliceFlag{
			Name: "use-pipe",
			Usage: fmt.Sprintf(`pipes for sink using:
			1. [logger] pipe is DEBUG logrus;
			2. [mongodb] pipe uses %s, %s and %s envs`,
				pipes.MongodbConnectURIEnvKey, pipes.MongodbDatabaseNameEnvKey, pipes.MongodbEnableJsonAttachEnvKey),
			EnvVar: "USE_PIPE",
			Value:  &cli.StringSlice{},
		},
		cli.BoolFlag{
			Name:   "pipes-parallel",
			Usage:  "enable the pipes parallel",
			EnvVar: "PIPES_PARALLEL",
		},
	}

	app.Run(os.Args)
}

func appAction(c *cli.Context) {
	var (
		resyncPeriod  = c.Duration("resync-period")
		storageTTL    = c.Duration("storage-ttl")
		kubeconfigs   = c.StringSlice("kubeconfig")
		usePipes      = c.StringSlice("use-pipe")
		pipesParallel = c.Bool("pipes-parallel")

		stopChan = newSystemStopChannel()
		g        = &wait.Group{}
		kconfigs []*rest.Config
	)

	initLog(c)

	if len(kubeconfigs) == 0 {
		kconfig, err := rest.InClusterConfig()
		if err != nil {
			logrus.WithError(err).Fatalln("failed to create Kubernetes config from in-cluster")
		}

		kconfigs = append(kconfigs, kconfig)
	} else {
		for _, kconfigPath := range kubeconfigs {
			if len(kconfigPath) != 0 {
				if _, err := os.Stat(kconfigPath); err != nil {
					logrus.WithError(err).Fatalln("can't open", kconfigPath)
				} else {
					kconfig, err := clientcmd.BuildConfigFromFlags("", kconfigPath)
					if err != nil {
						logrus.WithError(err).Fatalln("failed to create Kubernetes config from", kconfigPath)
					}

					kconfigs = append(kconfigs, kconfig)
				}
			}
		}
	}

	for _, kconfig := range kconfigs {
		khost := kconfig.Host
		kclient, err := kubernetes.NewForConfig(kconfig)
		if err != nil {
			logrus.WithError(err).Fatalf("failed to create Kubernetes client for %s", khost)
		}

		g.StartWithChannel(
			stopChan,
			newEventExporter(
				kclient,
				khost,
				resyncPeriod,
				storageTTL,
				usePipes,
				pipesParallel,
			).Run,
		)
	}

	g.Wait()
}
