package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/juju/errors"
	normanLeader "github.com/rancher/norman/leader"
	normanSignal "github.com/rancher/norman/signal"
	"github.com/sirupsen/logrus"
	"github.com/thxcode/kubernetes-event-exporter/pkg/events/sinks/pipes"
	"github.com/thxcode/kubernetes-event-exporter/pkg/exporters"
	"github.com/thxcode/kubernetes-event-exporter/pkg/simplelogger"
	"github.com/thxcode/kubernetes-event-exporter/server"
	"github.com/urfave/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	k8sRest "k8s.io/client-go/rest"
	k8sClientcmd "k8s.io/client-go/tools/clientcmd"

	exporterApiContext "github.com/thxcode/kubernetes-event-exporter/pkg/api/context"
	huaweiControllers "github.com/thxcode/kubernetes-event-exporter/pkg/controllers/huawei"
)

var (
	Version = "dev"
)

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
		logrus.SetFormatter(simplelogger.NewSimpleFormatter())
	}

	logrus.SetOutput(os.Stdout)
}

func main() {
	app := cli.NewApp()
	app.Name = "kubernetes-event-exporter"
	app.Version = Version
	app.Usage = "An exporter exposes events of Kubernetes."
	app.Action = appAction

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "http-listen-port",
			Usage: "HTTP listen port",
			Value: 8080,
		},
		cli.IntFlag{
			Name:  "https-listen-port",
			Usage: "HTTPS listen port",
			Value: 8443,
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
		cli.StringFlag{
			Name:   "kubeconfig",
			Usage:  "kube config for accessing Kubernetes cluster",
			EnvVar: "KUBECONFIG",
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
			2. [mongodb] pipe uses %s, %s envs`,
				pipes.MongodbConnectURIEnvKey, pipes.MongodbDatabaseNameEnvKey),
			EnvVar: "USE_PIPE",
			Value:  &cli.StringSlice{},
		},
		cli.BoolFlag{
			Name:   "pipes-parallel",
			Usage:  "enable the pipes parallel",
			EnvVar: "PIPES_PARALLEL",
		},
		cli.DurationFlag{
			Name:   "termination-grace-period",
			Usage:  "period for termination",
			EnvVar: "TERMINATION_GRACE_PERIOD",
			Value:  1 * time.Minute,
		},
	}

	app.Run(os.Args)
}

func appAction(c *cli.Context) error {
	initLog(c)

	rootCtx := normanSignal.SigTermCancelContext(context.Background())

	var (
		kubeconfig = c.String("kubeconfig")
		kconfig    *k8sRest.Config
	)

	if len(kubeconfig) == 0 {
		config, err := k8sRest.InClusterConfig()
		if err != nil {
			return errors.Annotatef(err, "failed to create Kubernetes config from in-cluster")
		}

		kconfig = config
	} else {
		if _, err := os.Stat(kubeconfig); err != nil {
			logrus.WithError(err).Fatalln("can't open", kubeconfig)
		} else {
			config, err := k8sClientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				return errors.Annotatef(err, "failed to create Kubernetes config from", kubeconfig)
			}

			kconfig = config
		}
	}

	os.Unsetenv("KUBECONFIG")

	return appRun(rootCtx, kconfig, c.Int("http-listen-port"), c.Int("https-listen-port"), c.Duration("termination-grace-period"), &exporters.EventExporterConfig{
		ResyncPeriod:  c.Duration("resync-period"),
		StorageTTL:    c.Duration("storage-ttl"),
		UsePipes:      c.StringSlice("use-pipe"),
		PipesParallel: c.Bool("pipes-parallel"),
	})
}

func appRun(rootCtx context.Context, rancherBackendK8sConfig *k8sRest.Config, httpPort, httpsPort int, terminationGracePeriod time.Duration, eventExportConfigTemplate *exporters.EventExporterConfig) error {
	mgrContext, err := exporterApiContext.BuildScaledContext(rootCtx, rancherBackendK8sConfig, httpsPort)
	if err != nil {
		return errors.Annotate(err, "can't build scaled context")
	}

	if err := server.Start(rootCtx, httpPort, httpsPort, mgrContext); err != nil {
		return errors.Annotate(err, "can't start API schemas server")
	}

	stoppingChan := make(chan struct{})
	go normanLeader.RunOrDie(rootCtx, "huawei-event-controllers", mgrContext.K8sClient, func(ctx context.Context) {
		mgrContext.Leader = true

		ctx = context.WithValue(ctx, "stoppingChan", stoppingChan)
		huaweiControllers.Register(ctx, mgrContext, eventExportConfigTemplate)
		err := mgrContext.Start(ctx)
		if err != nil {
			panic(errors.Annotate(err, "can't start scaled context"))
		}

		logrus.Infof("Huawei event controllers startup complete")

		<-ctx.Done()
	})

	logrus.Infoln("running")
	<-rootCtx.Done()
	logrus.Infoln("stopping")
	select {
	case <-time.Tick(terminationGracePeriod):
		logrus.Warnln("stopped with timeout")
	case <-stoppingChan:
		logrus.Infoln("stopped")
	}

	return rootCtx.Err()
}
