package huaweistored

import (
	"context"
	"sync"

	normanStoreCrd "github.com/rancher/norman/store/crd"
	normanTypes "github.com/rancher/norman/types"
	typesApisHuaweiSchema "github.com/rancher/types/apis/cloud.huawei.com/v3/schema"
	typesClientHuawei "github.com/rancher/types/client/cloud/v3"
	typesConfig "github.com/rancher/types/config"
	exporterApiHuaweiStore "github.com/thxcode/kubernetes-event-exporter/pkg/api/store/huawei"

	// rancherCustomizationCluster "github.com/rancher/rancher/pkg/api/customization/cluster"
	// rancherStoreCluster "github.com/rancher/rancher/pkg/api/store/cluster"
	rancherStoreScoped "github.com/rancher/rancher/pkg/api/store/scoped"
)

func Setup(ctx context.Context, mgrContext *typesConfig.ScaledContext) error {
	schemas := mgrContext.Schemas

	factory := &normanStoreCrd.Factory{ClientGetter: mgrContext.ClientGetter}
	wg := &sync.WaitGroup{}

	createCrd(ctx, wg, factory, schemas, &typesApisHuaweiSchema.Version,
		typesClientHuawei.HuaWeiClusterEventType,
	)
	exporterApiHuaweiStore.CreateMongo(ctx, wg, schemas.Schema(&typesApisHuaweiSchema.Version, typesClientHuawei.HuaWeiEventLogType))

	wg.Wait()

	setupScopedTypes(schemas)

	return nil
}

func setupScopedTypes(schemas *normanTypes.Schemas) {
	for _, schema := range schemas.Schemas() {
		if schema.Scope != normanTypes.NamespaceScope || schema.Store == nil || schema.Store.Context() != typesConfig.ManagementStorageContext {
			continue
		}

		for _, key := range []string{"projectId", "clusterId"} {
			ns, ok := schema.ResourceFields["namespaceId"]
			if !ok {
				continue
			}

			if _, ok := schema.ResourceFields[key]; !ok {
				continue
			}

			schema.Store = rancherStoreScoped.NewScopedStore(key, schema.Store)
			ns.Required = false
			schema.ResourceFields["namespaceId"] = ns
			break
		}
	}
}
