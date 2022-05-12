/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package health

import (
	"context"
	"fmt"

	"github.com/aws/aws-virtual-kubelet/internal/metrics"

	"github.com/aws/aws-virtual-kubelet/internal/config"

	"github.com/aws/aws-virtual-kubelet/internal/vkvmaclient"

	vkvmagentv0 "github.com/aws/aws-virtual-kubelet/proto/vkvmagent/v0"
	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"
)

func CheckAppHealthOnce(ctx context.Context, pod *corev1.Pod) *checkResult {
	monitor := NewMonitor(pod, SubjectApp, "app.healthOnce", checkAppHealth)

	return monitor.check(ctx, monitor)
}

func checkAppHealth(ctx context.Context, m *Monitor) *checkResult {
	cfg := config.Config()

	pod := m.Resource.(*corev1.Pod)

	vkvmaClient := vkvmaclient.NewVkvmaClient(pod.Status.PodIP, cfg.VKVMAgentConnectionConfig.Port)

	appClient, err := vkvmaClient.GetApplicationLifecycleClient(context.Background())
	if err != nil {
		message := fmt.Sprintf("Can't get Application Lifecycle client for pod %v(%v): %v",
			pod.Name, pod.Namespace, err)
		result := NewCheckResult(m, true, message, nil)
		klog.Warningf("Premature check failure: %+v", result)
		metrics.GRPCAppClientErrors.Inc()
		return result
	}

	appHealthResp, err := appClient.CheckApplicationHealth(ctx, &vkvmagentv0.ApplicationHealthRequest{})
	if err != nil {
		message := fmt.Sprintf("Can't Check Application Lifecycle Health for pod %v(%v): %v",
			pod.Name, pod.Namespace, err)
		result := NewCheckResult(m, true, message, nil)
		klog.Warningf("Premature check failure: %+v", result)
		metrics.CheckApplicationHealthErrors.Inc()
		return result
	}
	klog.V(1).InfoS("Received application health response", "response", appHealthResp, "pod", klog.KObj(pod))

	return NewCheckResult(m, false, "application health check succeeded", nil)
}

func checkVkvmaConnection(ctx context.Context, m *Monitor) *checkResult {
	return NewCheckResult(m, false, "⚠️  TODO implement this monitor", nil)
}

func checkVkvmaHealth(ctx context.Context, m *Monitor) *checkResult {
	return NewCheckResult(m, false, "⚠️  TODO implement this monitor", nil)
}
