package huawei

import (
	"context"

	typesConfig "github.com/rancher/types/config"
	exporterEvent "github.com/thxcode/kubernetes-event-exporter/pkg/controllers/huawei/events"
	"github.com/thxcode/kubernetes-event-exporter/pkg/exporters"
)

func Register(ctx context.Context, mgrContext *typesConfig.ScaledContext, eventExportConfigTemplate *exporters.EventExporterConfig) {
	exporterEvent.Register(ctx, mgrContext, eventExportConfigTemplate)
}
