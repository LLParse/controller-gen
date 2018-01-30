package main

import (
	"flag"
	"path/filepath"

	"k8s.io/gengo/args"

	"github.com/golang/glog"
	customargs "github.com/llparse/controller-gen/args"
	"github.com/llparse/controller-gen/generators"
	"github.com/spf13/pflag"
)

func main() {
	genericArgs, customArgs := customargs.NewDefaults()

	genericArgs.GoHeaderFilePath = filepath.Join(args.DefaultSourceTree(), "k8s.io/kubernetes/hack/boilerplate/boilerplate.go.txt")
	genericArgs.OutputPackagePath = "k8s.io/kubernetes/pkg/generated/controller"

	genericArgs.AddFlags(pflag.CommandLine)
	customArgs.AddFlags(pflag.CommandLine, "k8s.io/api", "k8s.io/client-go/informers", "k8s.io/client-go/listers")

	flag.Set("logtostderr", "true")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	glog.V(1).Infof("CustomArgs: %+v", customArgs)

	// add group version package as input dirs for gengo
	for _, pkg := range customArgs.Groups {
		for _, v := range pkg.Versions {
			genericArgs.InputDirs = append(genericArgs.InputDirs, v.Package)
		}
	}
	glog.V(1).Infof("InputDirs: %+v", genericArgs.InputDirs)

	// Run it.
	if err := genericArgs.Execute(
		generators.NameSystems(),
		generators.DefaultNameSystem(),
		generators.Packages,
	); err != nil {
		glog.Fatalf("Error: %v", err)
	}
	glog.V(2).Info("Completed successfully.")
}
