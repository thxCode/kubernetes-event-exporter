package events

import (
	"context"
	"sync"

	typesConfig "github.com/rancher/types/config"
	"github.com/thxcode/kubernetes-event-exporter/pkg/exporters"

	k8sApisMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var once = &sync.Once{}

func Register(ctx context.Context, mgrContext *typesConfig.ScaledContext, eventExportConfigTemplate *exporters.EventExporterConfig) {
	lifecycle := &eventLifecycle{
		mgrContext,
		eventExportConfigTemplate,
		make(map[string]*exporters.EventExporter, 8),
		mgrContext.HuaWei.HuaWeiClusterEvents(k8sApisMetaV1.NamespaceAll),
		mgrContext.HuaWei.HuaWeiClusterEvents(k8sApisMetaV1.NamespaceAll).Controller().Lister(),
	}

	once.Do(func() {
		go func() {
			stoppingChan := ctx.Value("stoppingChan").(chan struct{})

			<-ctx.Done()
			lifecycle.terminateEventExporters()
			close(stoppingChan)
		}()
	})

	mgrContext.Management.Clusters(k8sApisMetaV1.NamespaceAll).AddHandler("mgmt-project-event-exporter-create", lifecycle.Sync)

	mgrContext.Management.Clusters(k8sApisMetaV1.NamespaceAll).AddLifecycle("mgmt-cluster-event-exporter-remove", lifecycle)
}
