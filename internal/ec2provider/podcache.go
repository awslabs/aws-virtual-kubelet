/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package ec2provider

import (
	"errors"
	"fmt"
	"sync"

	"k8s.io/klog/v2"

	"github.com/aws/aws-virtual-kubelet/internal/utils"

	corev1 "k8s.io/api/core/v1"
)

type PodCache struct {
	pods map[string]*MetaPod
	lock sync.RWMutex
}

// GetList returns a list of Pods
func (pc *PodCache) GetList() []*MetaPod {
	pc.lock.RLock()
	defer pc.lock.RUnlock()

	var list []*MetaPod

	for _, value := range pc.pods {
		list = append(list, value)
	}

	return list
}

// GetPodList returns a list of Pods
func (pc *PodCache) GetPodList() []*corev1.Pod {
	pc.lock.RLock()
	defer pc.lock.RUnlock()

	var podList []*corev1.Pod

	for _, value := range pc.pods {
		podList = append(podList, value.pod)
	}

	return podList
}

// Get returns a value from the map given a key
func (pc *PodCache) Get(key string) *MetaPod {
	pc.lock.RLock()
	defer pc.lock.RUnlock()
	return pc.pods[key]
}

// Set updates the map element that has given key with the provided value
func (pc *PodCache) Set(key string, val *MetaPod) {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	pc.pods[key] = val
}

// UpdatePod updates the pod in the cached object (MetaPod) that has given key
func (pc *PodCache) UpdatePod(key string, pod *corev1.Pod) error {
	metaPod := pc.Get(key)

	if metaPod == nil {
		return errors.New(fmt.Sprintf("Can't find cache member with key %v to update", key))
	}

	metaPod.pod = pod
	pc.pods[key] = metaPod

	return nil
}

// Delete removes the map element with the given key
func (pc *PodCache) Delete(key string) {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	delete(pc.pods, key)
}

// CreateCacheFromPodList creates a new pod cache from a k8s PodList (empty if PodList is empty)
func CreateCacheFromPodList(podList *corev1.PodList) *PodCache {
	klog.InfoS("Rebuilding cache from pod list", "pods", len(podList.Items))

	podCache := &PodCache{
		pods: map[string]*MetaPod{},
		lock: sync.RWMutex{},
	}

	for _, pod := range podList.Items {
		klog.V(1).InfoS("Rehydrating pod for cache", "pod", klog.KObj(&pod), "annotations", &pod.Annotations)

		metaPod := NewMetaPod(pod.DeepCopy(), nil, nil) // monitors for pods are created elsewhere

		podKey := utils.GetPodCacheKey(pod.Namespace, pod.Name)
		podCache.Set(podKey, metaPod)
	}

	return podCache
}
