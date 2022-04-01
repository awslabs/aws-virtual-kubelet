package k8sutils

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

// K8SAPI defines the interface for K8S API actions for either Mock or official AWS Client usage.
type K8SAPI interface {
	// GetPods Retrieves all pods managed by VK
	GetPods(ctx context.Context, nodeName string) (*v1.PodList, error)
	// GetPod retrieves a Pod for the given PodName
	GetPod(ctx context.Context, namespace string, podName string) (*v1.Pod, error)
	// DeletePod deletes a pod from k8s cluster
	DeletePod(ctx context.Context, namespace string, podName string) error
}
