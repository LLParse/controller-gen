package generators

import (
	"io"

	clientgentypes "k8s.io/code-generator/cmd/client-gen/types"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
)

// controllerMainGenerator produces a controller main
type controllerMainGenerator struct {
	generator.DefaultGen
	controllerPackagePath string
	packagePath           string
	imports               namer.ImportTracker
	name                  string
	types                 []*types.Type
	groupVersionForType   map[*types.Type]clientgentypes.GroupVersion
}

var _ generator.Generator = &controllerMainGenerator{}

func (g *controllerMainGenerator) Namers(c *generator.Context) namer.NameSystems {
	pluralExceptions := map[string]string{
		"Endpoints": "Endpoints",
	}
	return namer.NameSystems{
		"raw":          namer.NewRawNamer(g.packagePath, g.imports),
		"publicPlural": namer.NewPublicPluralNamer(pluralExceptions),
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
		"types":                getResourceTypes(c, g.types, g.groupVersionForType),
		"Config":               c.Universe.Type(restConfig),
		"InClusterConfig":      c.Universe.Function(restInClusterConfigFunc),
		"BuildConfigFromFlags": c.Universe.Function(clientcmdBuildConfigFromFlagsFunc),
		"NewController":        c.Universe.Function(types.Name{Package: g.controllerPackagePath, Name: "NewController"}),
	}
	sw.Do(mainFunc, m)
	sw.Do(newKubeClientConfigFunc, m)
	sw.Do(makeStopChanFunc, m)
	return sw.Error()
}

var mainFunc = `
func main() {
  kubeconfig := flag.String("kubeconfig", "", "Path to a kube config; only required if out-of-cluster.")
  flag.Set("logtostderr", "true")
  flag.Parse()

  config, err := newKubeClientConfig(*kubeconfig)
  if err != nil {
    panic(err)
  }

  // FIXME make dynamic
  kubeClientset := kubernetes.NewForConfigOrDie(config)
  // FIXME make dynamic
  kubeInformerFactory := informers.NewSharedInformerFactory(kubeClientset, 0*time.Second)

  stopCh := makeStopChan()

  go $.NewController|raw$(
    // FIXME make dynamic
    kubeClientset,
    $- range .types$
    kubeInformerFactory.$.GroupName$().$.VersionName$().$.Type|publicPlural$(),
    $- end$
  ).Run(stopCh)

  // FIXME make dynamic
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
