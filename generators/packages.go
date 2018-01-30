package generators

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/golang/glog"

	"k8s.io/code-generator/cmd/client-gen/generators/util"
	clientgentypes "k8s.io/code-generator/cmd/client-gen/types"
	"k8s.io/gengo/args"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"

	customargs "github.com/llparse/controller-gen/args"
)

// NameSystems returns the name system used by the generators in this package.
func NameSystems() namer.NameSystems {
	pluralExceptions := map[string]string{
		"Endpoints": "Endpoints",
	}
	return namer.NameSystems{
		"public":             namer.NewPublicNamer(0),
		"private":            namer.NewPrivateNamer(0),
		"raw":                namer.NewRawNamer("", nil),
		"publicPlural":       namer.NewPublicPluralNamer(pluralExceptions),
		"allLowercasePlural": namer.NewAllLowercasePluralNamer(pluralExceptions),
		"lowercaseSingular":  &lowercaseSingularNamer{},
	}
}

// lowercaseSingularNamer implements Namer
type lowercaseSingularNamer struct{}

// Name returns t's name in all lowercase.
func (n *lowercaseSingularNamer) Name(t *types.Type) string {
	return strings.ToLower(t.Name.Name)
}

// DefaultNameSystem returns the default name system for ordering the types to be
// processed by the generators in this package.
func DefaultNameSystem() string {
	return "public"
}

// generatedBy returns information about the arguments used to invoke
// controller-gen.
func generatedBy() string {
	return fmt.Sprintf("\n// This file was automatically generated by controller-gen. DO NOT EDIT.\n\n")
}

// Packages makes the client package definition.
func Packages(context *generator.Context, arguments *args.GeneratorArgs) generator.Packages {
	boilerplate, err := arguments.LoadGoBoilerplate()
	if err != nil {
		glog.Fatalf("Failed loading boilerplate: %v", err)
	}

	customArgs, ok := arguments.CustomArgs.(*customargs.CustomArgs)
	if !ok {
		glog.Fatalf("cannot convert arguments.CustomArgs to customargs.CustomArgs")
	}

	groupVersionForType := make(map[*types.Type]clientgentypes.GroupVersion)

	var typesToGenerate []*types.Type
	for gv, types := range customArgs.Types {
		gvPackage := context.Universe.Package(filepath.Join(customArgs.ApiPackage, gv.Group.String(), gv.Version.String()))

		objectMeta, _, err := objectMetaForPackage(gvPackage)
		if err != nil {
			glog.Fatal(err)
		}
		if objectMeta == nil {
			// no types in this package had genclient
			continue
		}

		for _, typeName := range types {
			for packageTypeName, packageType := range gvPackage.Types {
				if strings.EqualFold(typeName, packageTypeName) {
					groupVersionForType[packageType] = gv
					typesToGenerate = append(typesToGenerate, packageType)
					// glog.V(3).Infof("%s: %+v", packageTypeName, packageType)
					break
				}
			}
		}
		if len(typesToGenerate) == 0 {
			continue
		}
		orderer := namer.Orderer{Namer: namer.NewPrivateNamer(0)}
		typesToGenerate = orderer.OrderTypes(typesToGenerate)
	}

	if len(typesToGenerate) == 0 {
		glog.Fatalf("no valid types were specified")
	}

	var packageList generator.Packages

	controllerPackagePath := filepath.Join(arguments.OutputPackagePath, "pkg", "controller", customArgs.Name)
	packageList = append(packageList, &generator.DefaultPackage{
		PackageName: customArgs.Name,
		PackagePath: controllerPackagePath,
		HeaderText:  append(boilerplate, []byte(generatedBy())...),
		GeneratorFunc: func(c *generator.Context) (generators []generator.Generator) {
			generators = append(generators, &controllerBaseGenerator{
				DefaultGen: generator.DefaultGen{
					OptionalName: "controller_base",
				},
				packagePath:         filepath.Join(arguments.OutputBase, controllerPackagePath),
				imports:             generator.NewImportTracker(),
				name:                customArgs.Name,
				types:               typesToGenerate,
				groupVersionForType: groupVersionForType,
				args:                customArgs,
			})
			return generators
		},
		FilterFunc: func(c *generator.Context, t *types.Type) bool {
			tags := util.MustParseClientGenTags(t.SecondClosestCommentLines)
			return tags.GenerateClient && tags.HasVerb("list") && tags.HasVerb("get")
		},
	})
	packageList = append(packageList, &generator.DefaultPackage{
		PackageName: customArgs.Name,
		PackagePath: controllerPackagePath,
		HeaderText:  boilerplate,
		GeneratorFunc: func(c *generator.Context) (generators []generator.Generator) {
			generators = append(generators, &controllerGenerator{
				DefaultGen: generator.DefaultGen{
					OptionalName: "controller",
				},
				packagePath:         filepath.Join(arguments.OutputBase, controllerPackagePath),
				imports:             generator.NewImportTracker(),
				name:                customArgs.Name,
				types:               typesToGenerate,
				groupVersionForType: groupVersionForType,
				args:                customArgs,
			})
			return generators
		},
		FilterFunc: func(c *generator.Context, t *types.Type) bool {
			tags := util.MustParseClientGenTags(t.SecondClosestCommentLines)
			return tags.GenerateClient && tags.HasVerb("list") && tags.HasVerb("get")
		},
	})

	controllerMainPackagePath := filepath.Join(arguments.OutputPackagePath, "cmd", "controller", customArgs.Name)
	packageList = append(packageList, &generator.DefaultPackage{
		PackageName: "main",
		PackagePath: controllerMainPackagePath,
		HeaderText:  append(boilerplate, []byte(generatedBy())...),
		GeneratorFunc: func(c *generator.Context) (generators []generator.Generator) {
			generators = append(generators, &controllerMainGenerator{
				DefaultGen: generator.DefaultGen{
					OptionalName: "main",
				},
				controllerPackagePath: controllerPackagePath,
				packagePath:           filepath.Join(arguments.OutputBase, controllerMainPackagePath),
				imports:               generator.NewImportTracker(),
				name:                  customArgs.Name,
				types:                 typesToGenerate,
				groupVersionForType:   groupVersionForType,
				args:                  customArgs,
			})
			return generators
		},
		FilterFunc: func(c *generator.Context, t *types.Type) bool {
			tags := util.MustParseClientGenTags(t.SecondClosestCommentLines)
			return tags.GenerateClient && tags.HasVerb("list") && tags.HasVerb("get")
		},
	})

	return packageList
}

// objectMetaForPackage returns the type of ObjectMeta used by package p.
func objectMetaForPackage(p *types.Package) (*types.Type, bool, error) {
	generatingForPackage := false
	for _, t := range p.Types {
		// filter out types which dont have genclient.
		if !util.MustParseClientGenTags(t.SecondClosestCommentLines).GenerateClient {
			continue
		}
		generatingForPackage = true
		for _, member := range t.Members {
			if member.Name == "ObjectMeta" {
				return member.Type, isInternal(member), nil
			}
		}
	}
	if generatingForPackage {
		return nil, false, fmt.Errorf("unable to find ObjectMeta for any types in package %s", p.Path)
	}
	return nil, false, nil
}

// isInternal returns true if the tags for a member do not contain a json tag
func isInternal(m types.Member) bool {
	return !strings.Contains(m.Tags, "json")
}
