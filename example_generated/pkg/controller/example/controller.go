// This file was automatically generated by controller-gen

package example

import (
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/wait"
	core_v1 "k8s.io/client-go/informers/core/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
	cache "k8s.io/client-go/tools/cache"
	workqueue "k8s.io/client-go/util/workqueue"
)

type Controller struct {
	kubeClient kubernetes.Interface

	podLister       v1.PodLister
	podListerSynced cache.InformerSynced

	podQueue workqueue.RateLimitingInterface
}

func NewController(
	kubeClient kubernetes.Interface,
	podInformer core_v1.PodInformer,
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
