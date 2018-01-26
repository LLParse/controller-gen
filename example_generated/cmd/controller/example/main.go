// This file was automatically generated by controller-gen

package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
	example "github.com/llparse/controller-gen/example_generated/pkg/controller/example"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
)

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

	go example.NewController(
		// FIXME make dynamic
		kubeClientset,
		// FIXME make dynamic
		kubeInformerFactory.Core().V1().Events(),
		kubeInformerFactory.Core().V1().Nodes(),
		kubeInformerFactory.Core().V1().Pods(),
		kubeInformerFactory.Core().V1().ReplicationControllers(),
		kubeInformerFactory.Core().V1().Services(),
	).Run(stopCh)

	// FIXME make dynamic
	kubeInformerFactory.Start(stopCh)

	<-stopCh
}

func newKubeClientConfig(configPath string) (*rest.Config, error) {
	if configPath != "" {
		return clientcmd.BuildConfigFromFlags("", configPath)
	}
	return rest.InClusterConfig()
}

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
