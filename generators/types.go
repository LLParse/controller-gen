package generators

import "k8s.io/gengo/types"

var (
	kubernetesInterface               = types.Name{Package: "k8s.io/client-go/kubernetes", Name: "Interface"}
	restConfig                        = types.Name{Package: "k8s.io/client-go/rest", Name: "Config"}
	restInClusterConfigFunc           = types.Name{Package: "k8s.io/client-go/rest", Name: "InClusterConfig"}
	cacheInformerSyncedFunc           = types.Name{Package: "k8s.io/client-go/tools/cache", Name: "InformerSynced"}
	clientcmdBuildConfigFromFlagsFunc = types.Name{Package: "k8s.io/client-go/tools/clientcmd", Name: "BuildConfigFromFlags"}
	workqueueRateLimitingInterface    = types.Name{Package: "k8s.io/client-go/util/workqueue", Name: "RateLimitingInterface"}
)

type ResourceType struct {
	Name        string
	Type        *types.Type
	GroupName   string
	VersionName string
	Informer    Informer
	Lister      Lister
	Queue       Queue
	Worker      Worker
}

type Informer struct {
	VariableName string
	Type         *types.Type
}

type Lister struct {
	VariableName string
	Type         *types.Type

	InformerSyncedVariableName string
	InformerSyncedFunction     *types.Type
}

type Queue struct {
	VariableName  string
	InterfaceType *types.Type
}

type Worker struct {
	VariableName string
}
