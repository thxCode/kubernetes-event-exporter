package schema

import (
	"net/http"

	"github.com/rancher/norman/types"
	m "github.com/rancher/norman/types/mapper"
	"github.com/rancher/types/apis/cloud.huawei.com/v3"
	"github.com/rancher/types/factory"
	"github.com/rancher/types/mapper"
	"k8s.io/api/core/v1"
)

var (
	Version = types.APIVersion{
		Version: "v3",
		Group:   "cloud.huawei.com",
		Path:    "/v3",
	}

	Schemas = factory.Schemas(&Version).
		Init(kubernetesEventTypes)
)

func kubernetesEventTypes(schema *types.Schemas) *types.Schemas {
	return schema.
		AddMapperForType(&Version, v1.ObjectReference{},
			&m.Drop{Field: "uid"},
			&m.Drop{Field: "apiVersion"},
			&m.Drop{Field: "resourceVersion"},
			&m.Move{From: "name", To: "resourceName"},
			&m.Move{From: "kind", To: "resourceKind"},
		).
		AddMapperForType(&Version, v3.HuaWeiEventLogFilter{},
			&mapper.NamespaceIDMapper{},
		).
		AddMapperForType(&Version, v3.HuaWeiClusterEvent{},
			&m.Embed{Field: "status"},
			m.DisplayName{},
		).
		AddMapperForType(&Version, v3.HuaWeiEventLog{},
			&m.Move{From: "type", To: "logType"},
			&m.Embed{Field: "involvedObject"},
		).
		MustImport(&Version, v1.ObjectReference{}).
		MustImport(&Version, v3.HuaWeiEventLogFilter{}).
		MustImportAndCustomize(&Version, v3.HuaWeiClusterEvent{}, func(schema *types.Schema) {
			schema.CollectionMethods = []string{http.MethodGet}
			schema.ResourceMethods = []string{http.MethodGet}

			// without fields
			delete(schema.ResourceFields, "creatorId")
			delete(schema.ResourceFields, "actions")

			// with filters
			newCollectionFilters := make(map[string]types.Filter, 2)
			newCollectionFilters["clusterId"] = types.Filter{
				Modifiers: []types.ModifierType{
					types.ModifierEQ,
				},
			}
			newCollectionFilters["isRemoved"] = types.Filter{
				Modifiers: []types.ModifierType{
					types.ModifierEQ,
				},
			}
			schema.CollectionFilters = newCollectionFilters

			// with action
			schema.ResourceActions["logs"] = types.Action{
				Input:  "huaWeiEventLogFilter",
				Output: "array[huaWeiEventLog]",
			}
		}).
		MustImportAndCustomize(&Version, v3.HuaWeiEventLog{}, func(schema *types.Schema) {
			schema.CollectionMethods = []string{http.MethodGet}
			schema.ResourceMethods = []string{http.MethodGet}

			// without fields
			delete(schema.ResourceFields, "creatorId")
			delete(schema.ResourceFields, "actions")

			// with filters
			newCollectionFilters := make(map[string]types.Filter, 7)
			newCollectionFilters["namespaceId"] = types.Filter{
				Modifiers: []types.ModifierType{
					types.ModifierEQ,
				},
			}
			newCollectionFilters["eventId"] = types.Filter{
				Modifiers: []types.ModifierType{
					types.ModifierEQ,
				},
			}
			newCollectionFilters["logType"] = types.Filter{
				Modifiers: []types.ModifierType{
					types.ModifierEQ,
				},
			}
			newCollectionFilters["resourceKind"] = types.Filter{
				Modifiers: []types.ModifierType{
					types.ModifierEQ,
				},
			}
			newCollectionFilters["createdRangeStart"] = types.Filter{
				Modifiers: []types.ModifierType{
					types.ModifierEQ,
				},
			}
			newCollectionFilters["createdRangeEnd"] = types.Filter{
				Modifiers: []types.ModifierType{
					types.ModifierEQ,
				},
			}
			newCollectionFilters["createdRangeFormat"] = types.Filter{
				Modifiers: []types.ModifierType{
					types.ModifierEQ,
				},
			}
			schema.CollectionFilters = newCollectionFilters

			// field modify, don't sort me
			// typeField := schema.ResourceFields["type"]
			// typeField.Options = []string{"All", "Normal", "Warning"}
			// typeField.Default = "All"
			// schema.ResourceFields["type"] = typeField
			//
			// kindField := schema.ResourceFields["kind"]
			// kindField.Options = []string{"All", "Node", "Pod", "Container"}
			// kindField.Default = "All"
			// schema.ResourceFields["kind"] = kindField

		})
}
