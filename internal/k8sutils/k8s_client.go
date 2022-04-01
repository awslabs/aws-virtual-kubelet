package k8sutils

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

type k8sClient struct {
	Svc typev1.CoreV1Interface
}

func NewK8sClient(configLocation string) (*k8sClient, error) {
	kubeconfig := filepath.Clean(configLocation)

	if _, err := os.Stat(kubeconfig); errors.Is(err, os.ErrNotExist) {
		klog.Warningf("kubeconfig file %v does not exist, using empty/default config", kubeconfig)
		kubeconfig = "" // setting this to empty will cause BuildConfigFromFlags to try InCluster then default configs
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &k8sClient{
		Svc: clientset.CoreV1(),
	}, nil
}

// GetPods retrieves a list of all pods running on the provider (can be cached).
func (client *k8sClient) GetPods(ctx context.Context, nodeName string) (*v1.PodList, error) {
	pods, err := client.Svc.Pods("").
		List(ctx, metav1.ListOptions{FieldSelector: "spec.nodeName=" + nodeName})
	if err != nil {
		return nil, err
	}
	return pods, nil
}

// GetPod retrieves a Pod for the given PodName
func (client *k8sClient) GetPod(ctx context.Context, namespace string, podName string) (*v1.Pod, error) {
	pod, err := client.Svc.Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return pod, nil
}

// DeletePod deletes the pod from k8s cluster
func (client *k8sClient) DeletePod(ctx context.Context, namespace string, podName string) error {
	return client.Svc.Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
}
