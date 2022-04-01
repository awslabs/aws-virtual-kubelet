/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package utils

import (
	"context"

	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPodCondition util method to create a podCondition based on provided conditionType
func GetPodCondition(conditionType corev1.PodConditionType, conditionStatus corev1.ConditionStatus,
	message string) corev1.PodCondition {
	return corev1.PodCondition{
		Type:               conditionType,
		Status:             conditionStatus,
		LastProbeTime:      metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Message:            message,
	}
}

func GetPodCacheKey(namespace string, name string) string {
	return namespace + "-" + name
}

func UpdateOrAppendPodCondition(pod *corev1.Pod, newCond corev1.PodCondition) {
	updated := false

	for i, condition := range pod.Status.Conditions {
		if condition.Type == newCond.Type {
			pod.Status.Conditions[i].Status = newCond.Status
			pod.Status.Conditions[i].Message = newCond.Message
			updated = true
		}
	}

	if !updated {
		pod.Status.Conditions = append(pod.Status.Conditions, newCond)
	}

	//klog.Infof("Notifying VK of state change for pod %v(%v): %v", pod.Name, pod.Namespace, newCond)
	//podNotifier(pod)
}

func PodIsWarmPool(ctx context.Context, pod *corev1.Pod, configLength int) bool {
	if configLength > 0 {
		return true
	}
	return false
}

// ReplaceCompute removes a pod's compute and obtains new compute without changing other pod properties or state
func ReplaceCompute(ctx context.Context, pod *corev1.Pod) error {
	klog.InfoS("Attempting to replace compute...", "pod", klog.KObj(pod))
	klog.InfoS("🚧 TODO implement ec2provider compute replacement", "pod", klog.KObj(pod))
	return nil
}
