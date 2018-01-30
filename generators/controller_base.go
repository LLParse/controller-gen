package generators

import (
	"io"

	clientgentypes "k8s.io/code-generator/cmd/client-gen/types"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
)

// controllerBaseGenerator produces a controller main
type controllerBaseGenerator struct {
	generator.DefaultGen
	packagePath         string
	imports             namer.ImportTracker
	name                string
	types               []*types.Type
	groupVersionForType map[*types.Type]clientgentypes.GroupVersion
}

var _ generator.Generator = &controllerBaseGenerator{}

func (g *controllerBaseGenerator) Namers(c *generator.Context) namer.NameSystems {
	pluralExceptions := map[string]string{
		"Endpoints": "Endpoints",
	}
	return namer.NameSystems{
		"raw":          namer.NewRawNamer(g.packagePath, g.imports),
		"publicPlural": namer.NewPublicPluralNamer(pluralExceptions),
	}
}

// We only want to call GenerateType() once for all types
func (g *controllerBaseGenerator) Filter(c *generator.Context, t *types.Type) bool {
	return t == g.types[0]
}

func (g *controllerBaseGenerator) Imports(c *generator.Context) []string {
	return g.imports.ImportLines()
}

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

func (g *controllerBaseGenerator) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "$", "$")

	m := map[string]interface{}{
		"types":      getResourceTypes(c, g.types, g.groupVersionForType),
		"Name":       g.name,
		"KubeClient": c.Universe.Type(kubernetesInterface),
	}

	sw.Do(controllerType, m)
	sw.Do(newControllerFunc, m)
	sw.Do(controllerRunFunc, m)
	sw.Do(controllerEnqueueWorkFunc, m)
	sw.Do(controllerPodWorkerFunc, m)
	return sw.Error()
}

var controllerType = `
type Controller struct {
  // FIXME make dynamic
  kubeClient $.KubeClient|raw$

  $range .types$
  $.Lister.VariableName$ $.Lister.Type|raw$
  $.Lister.InformerSyncedVariableName$ $.Lister.InformerSyncedFunction|raw$
  $- end$

  $range .types$
  $.Queue.VariableName$ $.Queue.InterfaceType|raw$
  $- end$
}
`

var newControllerFunc = `
func NewController(
  // FIXME make dynamic
  kubeClient $.KubeClient|raw$,
  $- range .types$
  $.Informer.VariableName$ $.Informer.Type|raw$,
  $- end$
) *Controller {
  ctrl := &Controller{
    kubeClient: kubeClient,
    $- range .types$
    $.Queue.VariableName$: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "$.Name$"),
    $- end$
  }

  $range .types$
  $.Informer.VariableName$.Informer().AddEventHandler(
    cache.ResourceEventHandlerFuncs{
      AddFunc:    func(obj interface{}) { ctrl.enqueueWork(ctrl.$.Queue.VariableName$, obj) },
      UpdateFunc: func(oldObj, newObj interface{}) { ctrl.enqueueWork(ctrl.$.Queue.VariableName$, newObj) },
      DeleteFunc: func(obj interface{}) { ctrl.enqueueWork(ctrl.$.Queue.VariableName$, obj) },
    },
  )
  $- end$

  $range .types$
  ctrl.$.Lister.VariableName$ = $.Informer.VariableName$.Lister()
  ctrl.$.Lister.InformerSyncedVariableName$ = $.Informer.VariableName$.Informer().HasSynced
  $- end$

  return ctrl
}
`

var controllerRunFunc = `
func (ctrl *Controller) Run(stopCh <-chan struct{}) {
  $- range .types$
  defer ctrl.$.Queue.VariableName$.ShutDown()
  $- end$

  glog.Infof("Starting $.Name$ controller")
  defer glog.Infof("Shutting down $.Name$ Controller")

  if !cache.WaitForCacheSync(stopCh
    $- range .types -$
    , ctrl.$.Lister.InformerSyncedVariableName$
    $- end -$
    ) {
    return
  }

  $range .types$
  go wait.Until(ctrl.$.Worker.VariableName$, time.Second, stopCh)
  $- end$

  <-stopCh
}
`

var controllerEnqueueWorkFunc = `
func (ctrl *Controller) enqueueWork(queue workqueue.Interface, obj interface{}) {
  // Beware of "xxx deleted" events
  if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
    obj = unknown.Obj
  }
  objName, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
  if err != nil {
    glog.Errorf("failed to get key from object: %v", err)
    return
  }
  glog.V(5).Infof("enqueued %q for sync", objName)
  queue.Add(objName)
}
`

var controllerPodWorkerFunc = `
$range .types$
func (ctrl *Controller) $.Worker.VariableName$() {
  workFunc := func() bool {
    keyObj, quit := ctrl.$.Queue.VariableName$.Get()
    if quit {
      return true
    }
    defer ctrl.$.Queue.VariableName$.Done(keyObj)
    key := keyObj.(string)
    glog.V(5).Infof("$.Worker.VariableName$[%s]", key)

    // TODO actual work
    return false
  }
  for {
    if quit := workFunc(); quit {
      glog.Infof("$.Worker.VariableName$ queue shutting down")
      return
    }
  }
}
$- end$
`
