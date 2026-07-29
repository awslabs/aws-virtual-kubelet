// Copyright © 2017 The virtual-kubelet authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manager

import (
	"github.com/virtual-kubelet/virtual-kubelet/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	corev1listers "k8s.io/client-go/listers/core/v1"
)

// ResourceManager acts as a passthrough to a cache (lister) for pods assigned to the current node.
// It is also a passthrough to a cache (lister) for Kubernetes secrets and config maps.
type ResourceManager struct {
	podLister       corev1listers.PodLister
	secretLister    corev1listers.SecretLister
	configMapLister corev1listers.ConfigMapLister
	serviceLister   corev1listers.ServiceLister
	pvcLister       corev1listers.PersistentVolumeClaimLister
	pvLister        corev1listers.PersistentVolumeLister
}

// NewResourceManager returns a ResourceManager with the internal maps initialized.
func NewResourceManager(podLister corev1listers.PodLister,
	secretLister corev1listers.SecretLister,
	configMapLister corev1listers.ConfigMapLister,
	serviceLister corev1listers.ServiceLister,
	pvcLister corev1listers.PersistentVolumeClaimLister,
	pvLister corev1listers.PersistentVolumeLister) (*ResourceManager, error) {
	rm := ResourceManager{
		podLister:       podLister,
		secretLister:    secretLister,
		configMapLister: configMapLister,
		serviceLister:   serviceLister,
		pvcLister:       pvcLister,
		pvLister:        pvLister,
	}
	return &rm, nil
}

// GetPod retrieves the specified pod from Kubernetes.
func (rm *ResourceManager) GetPod(name, namespace string) (*v1.Pod, error) {
	return rm.podLister.Pods(namespace).Get(name)
}

// GetPods returns a list of all known pods assigned to this virtual node.
func (rm *ResourceManager) GetPods() []*v1.Pod {
	l, err := rm.podLister.List(labels.Everything())
	if err == nil {
		return l
	}
	log.L.Errorf("failed to fetch pods from lister: %v", err)
	return make([]*v1.Pod, 0)
}

// GetConfigMap retrieves the specified config map from the cache.
func (rm *ResourceManager) GetConfigMap(name, namespace string) (*v1.ConfigMap, error) {
	return rm.configMapLister.ConfigMaps(namespace).Get(name)
}

// GetSecret retrieves the specified secret from Kubernetes.
func (rm *ResourceManager) GetSecret(name, namespace string) (*v1.Secret, error) {
	return rm.secretLister.Secrets(namespace).Get(name)
}

// GetSecrets retrieves the all secrets of a namespace from Kubernetes.
func (rm *ResourceManager) GetSecrets(namespace string) ([]*v1.Secret, error) {
	return rm.secretLister.Secrets(namespace).List(labels.Everything())
}

// ListServices retrieves the list of services from Kubernetes.
func (rm *ResourceManager) ListServices() ([]*v1.Service, error) {
	return rm.serviceLister.List(labels.Everything())
}

// GetPersistentVolumeClaim retrieves the specified pvc from Kubernetes
func (rm *ResourceManager) GetPersistentVolumeClaim(name, namespace string) (*v1.PersistentVolumeClaim, error) {
	return rm.pvcLister.PersistentVolumeClaims(namespace).Get(name)
}

// GetPersistentVolume retrieves the specified pv from Kubernetes
func (rm *ResourceManager) GetPersistentVolume(name string) (*v1.PersistentVolume, error) {
	return rm.pvLister.Get(name)
}
