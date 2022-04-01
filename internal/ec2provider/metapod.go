/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package ec2provider

import (
	"github.com/aws/aws-virtual-kubelet/internal/health"
	corev1 "k8s.io/api/core/v1"
)

// 🐛

type MetaPod struct {
	pod      *corev1.Pod
	monitor  *health.PodMonitor
	notifier func(*corev1.Pod)
}

func NewMetaPod(pod *corev1.Pod, monitor *health.PodMonitor, notifier func(*corev1.Pod)) *MetaPod {
	return &MetaPod{
		pod:      pod,
		monitor:  monitor,
		notifier: notifier,
	}
}
