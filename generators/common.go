package generators

import (
	"path/filepath"
	"strings"

	clientgentypes "k8s.io/code-generator/cmd/client-gen/types"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"

	"github.com/llparse/controller-gen/args"
)

func getResourceTypes(c *generator.Context, rTypes []*types.Type, gvForType map[*types.Type]clientgentypes.GroupVersion,
	args *args.CustomArgs) (resourceTypes []ResourceType) {
	for _, t := range rTypes {
		nameLowerFirst := strings.ToLower(t.Name.Name[:1]) + t.Name.Name[1:]
		group := gvForType[t].Group.NonEmpty()
		version := gvForType[t].Version.NonEmpty()
		resourceTypes = append(resourceTypes, ResourceType{
			Name:        t.Name.Name,
			Type:        t,
			GroupName:   namer.IC(group),
			VersionName: namer.IC(version),
			Informer: Informer{
				VariableName: nameLowerFirst + "Informer",
				Type:         c.Universe.Type(types.Name{Package: filepath.Join(args.InformerPackage, group, version), Name: t.Name.Name + "Informer"}),
			},
			Lister: Lister{
				VariableName: nameLowerFirst + "Lister",
				Type:         c.Universe.Type(types.Name{Package: filepath.Join(args.ListerPackage, group, version), Name: t.Name.Name + "Lister"}),
				InformerSyncedVariableName: strings.ToLower(t.Name.Name) + "ListerSynced",
				InformerSyncedFunction:     c.Universe.Function(cacheInformerSyncedFunc),
			},
			Queue: Queue{
				VariableName:  nameLowerFirst + "Queue",
				InterfaceType: c.Universe.Type(workqueueRateLimitingInterface),
			},
			Worker: Worker{
				VariableName: nameLowerFirst + "Worker",
			},
		})
	}
	return
}
