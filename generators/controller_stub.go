package generators

import (
	"io"

	clientgentypes "k8s.io/code-generator/cmd/client-gen/types"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"

	"github.com/llparse/controller-gen/args"
)

// controllerGenerator produces a controller
type controllerGenerator struct {
	generator.DefaultGen
	packagePath         string
	imports             namer.ImportTracker
	name                string
	types               []*types.Type
	groupVersionForType map[*types.Type]clientgentypes.GroupVersion
	args                *args.CustomArgs
}

var _ generator.Generator = &controllerGenerator{}

func (g *controllerGenerator) Namers(c *generator.Context) namer.NameSystems {
	pluralExceptions := map[string]string{
		"Endpoints": "Endpoints",
	}
	return namer.NameSystems{
		"raw":          namer.NewRawNamer(g.packagePath, g.imports),
		"publicPlural": namer.NewPublicPluralNamer(pluralExceptions),
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
		"types": getResourceTypes(c, g.types, g.groupVersionForType, g.args),
	}

	sw.Do(controllerPodWorkerFunc, m)
	return sw.Error()
}

var controllerPodWorkerFunc = `
$range .types$
// generated stub
func (ctrl *Controller) $.Worker.VariableName$() {
  workFunc := func() bool {
    keyObj, quit := ctrl.$.Queue.VariableName$.Get()
    if quit {
      return true
    }
    defer ctrl.$.Queue.VariableName$.Done(keyObj)
    key := keyObj.(string)
    glog.V(5).Infof("$.Worker.VariableName$[%s]", key)

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
