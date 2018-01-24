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

// controllerMainGenerator produces a controller main
type controllerMainGenerator struct {
	generator.DefaultGen
	packagePath string
	imports     namer.ImportTracker
	types       []*types.Type
}

// We only want to call GenerateType() once per group.
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
	}
	sw.Do(mainTemplate, m)
	sw.Do(newKubeClientConfig, m)
	sw.Do(makeStopChan, m)
	return sw.Error()
}

// k8s.io/client-go/informers, k8s.io/client-go/kubernetes
var mainTemplate = `
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

  go controller.NewController(
    kubeClientset,
    kubeInformerFactory.Core().V1().Nodes(),
    kubeInformerFactory.Core().V1().Pods(),
  ).Run(stopCh)

  kubeInformerFactory.Start(stopCh)

  <-stopCh
}
`

var newKubeClientConfig = `
func newKubeClientConfig(configPath string) (*$.Config|raw$, error) {
  if configPath != "" {
    return $.BuildConfigFromFlags|raw$("", configPath)
  }
  return $.InClusterConfig|raw$()
}
`

var makeStopChan = `
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
