package server

import (
	"context"
	"net/http"

	normanApi "github.com/rancher/norman/api"
	normanBuiltin "github.com/rancher/norman/api/builtin"
	normanSubscribe "github.com/rancher/norman/pkg/subscribe"
	rancherDynamicSchemaController "github.com/rancher/rancher/pkg/api/controllers/dynamicschema"
	rancherSettingController "github.com/rancher/rancher/pkg/api/controllers/settings"
	rancherManagementStored "github.com/rancher/rancher/pkg/api/server/managementstored"
	rancherUserStored "github.com/rancher/rancher/pkg/api/server/userstored"
	typesApisHuaweiSchema "github.com/rancher/types/apis/cloud.huawei.com/v3/schema"
	typesApisClusterSchema "github.com/rancher/types/apis/cluster.cattle.io/v3/schema"
	typesApisManagementSchema "github.com/rancher/types/apis/management.cattle.io/v3/schema"
	typesApisProjectSchema "github.com/rancher/types/apis/project.cattle.io/v3/schema"
	typesConfig "github.com/rancher/types/config"
	exporterHuaWeiStored "github.com/thxcode/kubernetes-event-exporter/pkg/api/server/huaweistored"
)

func New(ctx context.Context, mgrContext *typesConfig.ScaledContext, k8sProxy http.Handler) (http.Handler, error) {
	normanSubscribe.Register(&normanBuiltin.Version, mgrContext.Schemas)
	normanSubscribe.Register(&typesApisHuaweiSchema.Version, mgrContext.Schemas)
	normanSubscribe.Register(&typesApisManagementSchema.Version, mgrContext.Schemas)
	normanSubscribe.Register(&typesApisClusterSchema.Version, mgrContext.Schemas)
	normanSubscribe.Register(&typesApisProjectSchema.Version, mgrContext.Schemas)

	if err := exporterHuaWeiStored.Setup(ctx, mgrContext); err != nil {
		return nil, err
	}

	if err := rancherManagementStored.Setup(ctx, mgrContext, nil, k8sProxy); err != nil {
		return nil, err
	}

	if err := rancherUserStored.Setup(ctx, mgrContext, nil, k8sProxy); err != nil {
		return nil, err
	}

	server := normanApi.NewAPIServer()
	server.AccessControl = mgrContext.AccessControl

	if err := server.AddSchemas(mgrContext.Schemas); err != nil {
		return nil, err
	}

	rancherDynamicSchemaController.Register(mgrContext, server.Schemas)

	return server, rancherSettingController.Register(mgrContext)
}
