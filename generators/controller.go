package generators

import (
	"io"
	// "os"
	// "path/filepath"
	// "strings"

	// "k8s.io/code-generator/cmd/client-gen/generators/util"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
)

// controllerGenerator produces a controller main
type controllerGenerator struct {
	generator.DefaultGen
	packagePath string
	imports     namer.ImportTracker
	name        string
	types       []*types.Type
}

var _ generator.Generator = &controllerGenerator{}

func (g *controllerGenerator) Namers(c *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		"raw": namer.NewRawNamer(g.packagePath, g.imports),
	}
}

// We only want to call GenerateType() once for all types
func (g *controllerGenerator) Filter(c *generator.Context, t *types.Type) bool {
	return t == g.types[0]
}

func (g *controllerGenerator) Imports(c *generator.Context) []string {
	return g.imports.ImportLines()
}

func (g *controllerGenerator) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "$", "$")
	m := map[string]interface{}{
		"PodLister":   c.Universe.Type(types.Name{Package: "k8s.io/client-go/listers/core/v1", Name: "PodLister"}),
		"PodInformer": c.Universe.Type(types.Name{Package: "k8s.io/client-go/informers/core/v1", Name: "PodInformer"}),
	}
	sw.Do(controllerType, m)
	sw.Do(newControllerFunc, m)
	sw.Do(controllerRunFunc, m)
	sw.Do(controllerEnqueueWorkFunc, m)
	sw.Do(controllerPodWorkerFunc, m)
	return sw.Error()
}

// listerscorev1 "k8s.io/client-go/listers/core/v1", "k8s.io/client-go/tools/cache",
// "k8s.io/client-go/kubernetes", "k8s.io/client-go/util/workqueue"
var controllerType = `
type Controller struct {
  kubeClient kubernetes.Interface

  podLister       $.PodLister|raw$
  podListerSynced cache.InformerSynced

  podQueue workqueue.RateLimitingInterface
}
`

// informerscorev1 "k8s.io/client-go/informers/core/v1", "k8s.io/client-go/kubernetes",
// "k8s.io/client-go/util/workqueue"
var newControllerFunc = `
func NewController(
  kubeClient kubernetes.Interface,
  podInformer $.PodInformer|raw$,
) *Controller {
  ctrl := &Controller{
    kubeClient: kubeClient,
    podQueue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pod"),
  }

  podInformer.Informer().AddEventHandler(
    cache.ResourceEventHandlerFuncs{
      AddFunc:    func(obj interface{}) { ctrl.enqueueWork(ctrl.podQueue, obj) },
      UpdateFunc: func(oldObj, newObj interface{}) { ctrl.enqueueWork(ctrl.podQueue, newObj) },
      DeleteFunc: func(obj interface{}) { ctrl.enqueueWork(ctrl.podQueue, obj) },
    },
  )

  ctrl.podLister = podInformer.Lister()
  ctrl.podListerSynced = podInformer.Informer().HasSynced

  return ctrl
}
`

var controllerRunFunc = `
func (ctrl *Controller) Run(stopCh <-chan struct{}) {
  defer ctrl.podQueue.ShutDown()

  glog.Infof("Starting pod controller")
  defer glog.Infof("Shutting down pod Controller")

  if !cache.WaitForCacheSync(stopCh, ctrl.podListerSynced) {
    return
  }

  go wait.Until(ctrl.podWorker, time.Second, stopCh)

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
func (ctrl *Controller) podWorker() {
  workFunc := func() bool {
    keyObj, quit := ctrl.podQueue.Get()
    if quit {
      return true
    }
    defer ctrl.podQueue.Done(keyObj)
    key := keyObj.(string)
    glog.V(5).Infof("podWorker[%s]", key)
    return false
  }
  for {
    if quit := workFunc(); quit {
      glog.Infof("pod worker queue shutting down")
      return
    }
  }
}
`

// controllerMainGenerator produces a controller main
type controllerMainGenerator struct {
	generator.DefaultGen
	controllerPackagePath string
	packagePath           string
	imports               namer.ImportTracker
	name                  string
	types                 []*types.Type
}

var _ generator.Generator = &controllerMainGenerator{}

func (g *controllerMainGenerator) Namers(c *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		"raw": namer.NewRawNamer(g.packagePath, g.imports),
	}
}

// We only want to call GenerateType() once for all types
func (g *controllerMainGenerator) Filter(c *generator.Context, t *types.Type) bool {
	return t == g.types[0]
}

func (g *controllerMainGenerator) Imports(c *generator.Context) []string {
	return g.imports.ImportLines()
}

func (g *controllerMainGenerator) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "$", "$")

	m := map[string]interface{}{
		"Config":               c.Universe.Type(types.Name{Package: "k8s.io/client-go/rest", Name: "Config"}),
		"InClusterConfig":      c.Universe.Function(types.Name{Package: "k8s.io/client-go/rest", Name: "InClusterConfig"}),
		"BuildConfigFromFlags": c.Universe.Function(types.Name{Package: "k8s.io/client-go/tools/clientcmd", Name: "BuildConfigFromFlags"}),
		"NewController":        c.Universe.Function(types.Name{Package: g.controllerPackagePath, Name: "NewController"}),
	}
	sw.Do(mainFunc, m)
	sw.Do(newKubeClientConfigFunc, m)
	sw.Do(makeStopChanFunc, m)
	return sw.Error()
}

// k8s.io/client-go/informers, k8s.io/client-go/kubernetes
var mainFunc = `
func main() {
  kubeconfig := flag.String("kubeconfig", "", "Path to a kube config; only required if out-of-cluster.")
  flag.Set("logtostderr", "true")
  flag.Parse()

  config, err := newKubeClientConfig(*kubeconfig)
  if err != nil {
    panic(err)
  }

  kubeClientset := kubernetes.NewForConfigOrDie(config)
  kubeInformerFactory := informers.NewSharedInformerFactory(kubeClientset, 0*time.Second)

  stopCh := makeStopChan()

  go $.NewController|raw$(
    kubeClientset,
    kubeInformerFactory.Core().V1().Pods(),
  ).Run(stopCh)

  kubeInformerFactory.Start(stopCh)

  <-stopCh
}
`

var newKubeClientConfigFunc = `
func newKubeClientConfig(configPath string) (*$.Config|raw$, error) {
  if configPath != "" {
    return $.BuildConfigFromFlags|raw$("", configPath)
  }
  return $.InClusterConfig|raw$()
}
`

var makeStopChanFunc = `
func makeStopChan() <-chan struct{} {
  stop := make(chan struct{})
  c := make(chan os.Signal, 2)
  signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

  go func() {
    <-c
    glog.Info("Received stop signal, attempting graceful termination...")
    close(stop)
    <-c
    glog.Info("Received stop signal, terminating immediately!")
    os.Exit(1)
  }()
  return stop
}
`
