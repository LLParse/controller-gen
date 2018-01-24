package generators

import (
  "io"
  "os"
  "path/filepath"
  "strings"

  "k8s.io/code-generator/cmd/client-gen/generators/util"
  "k8s.io/gengo/generator"
  "k8s.io/gengo/types"
)

// expansionGenerator produces a file for a expansion interfaces.
type expansionGenerator struct {
  generator.DefaultGen
  packagePath string
  types       []*types.Type
}

// We only want to call GenerateType() once per group.
func (g *expansionGenerator) Filter(c *generator.Context, t *types.Type) bool {
  return t == g.types[0]
}

func (g *expansionGenerator) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
  sw := generator.NewSnippetWriter(w, c, "$", "$")
  for _, t := range g.types {
    tags := util.MustParseClientGenTags(t.SecondClosestCommentLines)
    if _, err := os.Stat(filepath.Join(g.packagePath, strings.ToLower(t.Name.Name+"_expansion.go"))); os.IsNotExist(err) {
      sw.Do(expansionInterfaceTemplate, t)
      if !tags.NonNamespaced {
        sw.Do(namespacedExpansionInterfaceTemplate, t)
      }
    }
  }
  return sw.Error()
}

var expansionInterfaceTemplate = `
// $.|public$ListerExpansion allows custom methods to be added to
// $.|public$Lister.
type $.|public$ListerExpansion interface {}
`

var namespacedExpansionInterfaceTemplate = `
// $.|public$NamespaceListerExpansion allows custom methods to be added to
// $.|public$NamespaceLister.
type $.|public$NamespaceListerExpansion interface {}
`
