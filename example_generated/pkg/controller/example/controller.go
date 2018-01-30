package example

import "github.com/golang/glog"

// generated stub
func (ctrl *Controller) deploymentWorker() {
	workFunc := func() bool {
		keyObj, quit := ctrl.deploymentQueue.Get()
		if quit {
			return true
		}
		defer ctrl.deploymentQueue.Done(keyObj)
		key := keyObj.(string)
		glog.V(5).Infof("deploymentWorker[%s]", key)

		return false
	}
	for {
		if quit := workFunc(); quit {
			glog.Infof("deploymentWorker queue shutting down")
			return
		}
	}
}

// generated stub
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

// generated stub
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

// generated stub
func (ctrl *Controller) storageClassWorker() {
	workFunc := func() bool {
		keyObj, quit := ctrl.storageClassQueue.Get()
		if quit {
			return true
		}
		defer ctrl.storageClassQueue.Done(keyObj)
		key := keyObj.(string)
		glog.V(5).Infof("storageClassWorker[%s]", key)

		return false
	}
	for {
		if quit := workFunc(); quit {
			glog.Infof("storageClassWorker queue shutting down")
			return
		}
	}
}
