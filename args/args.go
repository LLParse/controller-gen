package args

import (
	"path"

	"github.com/spf13/pflag"
	clientgenTypes "k8s.io/code-generator/cmd/client-gen/types"
	codegenutil "k8s.io/code-generator/pkg/util"
	"k8s.io/gengo/args"
)

var DefaultInputDirs = []string{}

type CustomArgs struct {
	// Name of the controller to generate.
	Name string

	// A sorted list of group versions to generate. For each of them the package path is found
	// in GroupVersionToInputPath.
	Groups []clientgenTypes.GroupVersions

	// Overrides for which types should be included in the client.
	Types map[clientgenTypes.GroupVersion][]string

	BasePath string
}

func NewDefaults() (*args.GeneratorArgs, *CustomArgs) {
	genericArgs := args.Default().WithoutDefaultFlagParsing()
	customArgs := &CustomArgs{
		Name: "example",
	}
	genericArgs.CustomArgs = customArgs
	genericArgs.InputDirs = DefaultInputDirs

	if pkg := codegenutil.CurrentPackage(); len(pkg) != 0 {
		genericArgs.OutputPackagePath = path.Join(pkg, "generated/controllers")
	}

	return genericArgs, customArgs
}

func (ca *CustomArgs) AddFlags(fs *pflag.FlagSet, inputBase string) {
	gvsBuilder := NewGroupVersionsBuilder(&ca.Groups)
	pflag.Var(NewInputBasePathValue(gvsBuilder, &ca.BasePath, inputBase), "api-package", "base path to look for the api group.")
	pflag.Var(NewGVTypesValue(gvsBuilder, &ca.Types, []string{}), "types", "list of group/version/type for which controller should receive change events.")
	pflag.StringVarP(&ca.Name, "name", "n", ca.Name, "the name of the generated controller.")
}
