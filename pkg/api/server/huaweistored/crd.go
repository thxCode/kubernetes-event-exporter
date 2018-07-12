package huaweistored

import (
	"context"
	"sync"

	normanStoreCrd "github.com/rancher/norman/store/crd"
	normanTypes "github.com/rancher/norman/types"
	typesConfig "github.com/rancher/types/config"
)

func createCrd(ctx context.Context, wg *sync.WaitGroup, factory *normanStoreCrd.Factory, schemas *normanTypes.Schemas, version *normanTypes.APIVersion, schemaIDs ...string) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		var schemasToCreate []*normanTypes.Schema

		for _, schemaID := range schemaIDs {
			s := schemas.Schema(version, schemaID)
			if s == nil {
				panic("can not find schema " + schemaID)
			}
			schemasToCreate = append(schemasToCreate, s)
		}

		err := factory.AssignStores(ctx, typesConfig.ManagementStorageContext, schemasToCreate...)
		if err != nil {
			panic("creating CRD store " + err.Error())
		}
	}()
}