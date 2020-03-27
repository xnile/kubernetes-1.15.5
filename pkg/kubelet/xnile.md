
kuberuntime
container_runtime

dockershim


kubelet.Run->kubelet.syncLoop()->kubelet.syncLoopIteration()->kubelet.syncPod()->SyncHandler->kubelet.dispatchWork()->podWorkers.managePodLoop()->kubelet.syncPod()->kubeGenericRuntimeManager.SyncPod()->