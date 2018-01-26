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
	// FIXME make dynamic
	kubeClient kubernetes.Interface

	eventLister                       v1.EventLister
	eventListerSynced                 cache.InformerSynced
	nodeLister                        v1.NodeLister
	nodeListerSynced                  cache.InformerSynced
	podLister                         v1.PodLister
	podListerSynced                   cache.InformerSynced
	replicationControllerLister       v1.ReplicationControllerLister
	replicationcontrollerListerSynced cache.InformerSynced
	serviceLister                     v1.ServiceLister
	serviceListerSynced               cache.InformerSynced

	eventQueue                 workqueue.RateLimitingInterface
	nodeQueue                  workqueue.RateLimitingInterface
	podQueue                   workqueue.RateLimitingInterface
	replicationControllerQueue workqueue.RateLimitingInterface
	serviceQueue               workqueue.RateLimitingInterface
}

func NewController(
	// FIXME make dynamic
	kubeClient kubernetes.Interface,
	eventInformer core_v1.EventInformer,
	nodeInformer core_v1.NodeInformer,
	podInformer core_v1.PodInformer,
	replicationControllerInformer core_v1.ReplicationControllerInformer,
	serviceInformer core_v1.ServiceInformer,
) *Controller {
	ctrl := &Controller{
		kubeClient:                 kubeClient,
		eventQueue:                 workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Event"),
		nodeQueue:                  workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Node"),
		podQueue:                   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Pod"),
		replicationControllerQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ReplicationController"),
		serviceQueue:               workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Service"),
	}

	eventInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { ctrl.enqueueWork(ctrl.eventQueue, obj) },
			UpdateFunc: func(oldObj, newObj interface{}) { ctrl.enqueueWork(ctrl.eventQueue, newObj) },
			DeleteFunc: func(obj interface{}) { ctrl.enqueueWork(ctrl.eventQueue, obj) },
		},
	)
	nodeInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { ctrl.enqueueWork(ctrl.nodeQueue, obj) },
			UpdateFunc: func(oldObj, newObj interface{}) { ctrl.enqueueWork(ctrl.nodeQueue, newObj) },
			DeleteFunc: func(obj interface{}) { ctrl.enqueueWork(ctrl.nodeQueue, obj) },
		},
	)
	podInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { ctrl.enqueueWork(ctrl.podQueue, obj) },
			UpdateFunc: func(oldObj, newObj interface{}) { ctrl.enqueueWork(ctrl.podQueue, newObj) },
			DeleteFunc: func(obj interface{}) { ctrl.enqueueWork(ctrl.podQueue, obj) },
		},
	)
	replicationControllerInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { ctrl.enqueueWork(ctrl.replicationControllerQueue, obj) },
			UpdateFunc: func(oldObj, newObj interface{}) { ctrl.enqueueWork(ctrl.replicationControllerQueue, newObj) },
			DeleteFunc: func(obj interface{}) { ctrl.enqueueWork(ctrl.replicationControllerQueue, obj) },
		},
	)
	serviceInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { ctrl.enqueueWork(ctrl.serviceQueue, obj) },
			UpdateFunc: func(oldObj, newObj interface{}) { ctrl.enqueueWork(ctrl.serviceQueue, newObj) },
			DeleteFunc: func(obj interface{}) { ctrl.enqueueWork(ctrl.serviceQueue, obj) },
		},
	)

	ctrl.eventLister = eventInformer.Lister()
	ctrl.eventListerSynced = eventInformer.Informer().HasSynced
	ctrl.nodeLister = nodeInformer.Lister()
	ctrl.nodeListerSynced = nodeInformer.Informer().HasSynced
	ctrl.podLister = podInformer.Lister()
	ctrl.podListerSynced = podInformer.Informer().HasSynced
	ctrl.replicationControllerLister = replicationControllerInformer.Lister()
	ctrl.replicationcontrollerListerSynced = replicationControllerInformer.Informer().HasSynced
	ctrl.serviceLister = serviceInformer.Lister()
	ctrl.serviceListerSynced = serviceInformer.Informer().HasSynced

	return ctrl
}

func (ctrl *Controller) Run(stopCh <-chan struct{}) {
	defer ctrl.eventQueue.ShutDown()
	defer ctrl.nodeQueue.ShutDown()
	defer ctrl.podQueue.ShutDown()
	defer ctrl.replicationControllerQueue.ShutDown()
	defer ctrl.serviceQueue.ShutDown()

	glog.Infof("Starting example controller")
	defer glog.Infof("Shutting down example Controller")

	if !cache.WaitForCacheSync(stopCh, ctrl.eventListerSynced, ctrl.nodeListerSynced, ctrl.podListerSynced, ctrl.replicationcontrollerListerSynced, ctrl.serviceListerSynced) {
		return
	}

	go wait.Until(ctrl.eventWorker, time.Second, stopCh)
	go wait.Until(ctrl.nodeWorker, time.Second, stopCh)
	go wait.Until(ctrl.podWorker, time.Second, stopCh)
	go wait.Until(ctrl.replicationControllerWorker, time.Second, stopCh)
	go wait.Until(ctrl.serviceWorker, time.Second, stopCh)

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

func (ctrl *Controller) eventWorker() {
	workFunc := func() bool {
		keyObj, quit := ctrl.eventQueue.Get()
		if quit {
			return true
		}
		defer ctrl.eventQueue.Done(keyObj)
		key := keyObj.(string)
		glog.V(5).Infof("eventWorker[%s]", key)
		return false
	}
	for {
		if quit := workFunc(); quit {
			glog.Infof("eventWorker queue shutting down")
			return
		}
	}
}
func (ctrl *Controller) nodeWorker() {
	workFunc := func() bool {
		keyObj, quit := ctrl.nodeQueue.Get()
		if quit {
			return true
		}
		defer ctrl.nodeQueue.Done(keyObj)
		key := keyObj.(string)
		glog.V(5).Infof("nodeWorker[%s]", key)
		return false
	}
	for {
		if quit := workFunc(); quit {
			glog.Infof("nodeWorker queue shutting down")
			return
		}
	}
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
			glog.Infof("podWorker queue shutting down")
			return
		}
	}
}
func (ctrl *Controller) replicationControllerWorker() {
	workFunc := func() bool {
		keyObj, quit := ctrl.replicationControllerQueue.Get()
		if quit {
			return true
		}
		defer ctrl.replicationControllerQueue.Done(keyObj)
		key := keyObj.(string)
		glog.V(5).Infof("replicationControllerWorker[%s]", key)
		return false
	}
	for {
		if quit := workFunc(); quit {
			glog.Infof("replicationControllerWorker queue shutting down")
			return
		}
	}
}
func (ctrl *Controller) serviceWorker() {
	workFunc := func() bool {
		keyObj, quit := ctrl.serviceQueue.Get()
		if quit {
			return true
		}
		defer ctrl.serviceQueue.Done(keyObj)
		key := keyObj.(string)
		glog.V(5).Infof("serviceWorker[%s]", key)
		return false
	}
	for {
		if quit := workFunc(); quit {
			glog.Infof("serviceWorker queue shutting down")
			return
		}
	}
}
