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

//func watchAppHealth(ctx context.Context, m *Monitor, ch *CheckHandler) error {
//	cfg := config.Config()
//
//	m.watcher.isWatching = true
//
//	pod := m.Resource.(*corev1.Pod)
//
//	klog.InfoS("Watching application health for pod", "pod", klog.KObj(pod))
//
//	vkvmaClient := vkvmaclient.NewVkvmaClient(pod.Status.PodIP, cfg.VKVMAgentConnectionConfig.Port)
//
//	appClient, err := vkvmaClient.GetApplicationLifecycleClient(context.Background())
//	if err != nil {
//		message := fmt.Sprintf("Can't get Application Lifecycle client for pod %v(%v): %v",
//			pod.Name, pod.Namespace, err)
//		result := NewCheckResult(m, true, message, nil)
//		klog.Warningf("Premature check failure: %+v", result)
//		metrics.GRPCAppClientErrors.Inc()
//		ch.in <- result
//		return nil
//	}
//
//	stream, err := appClient.WatchApplicationHealth(context.Background(), &vkvmagentv0.ApplicationHealthRequest{})
//	if err != nil {
//		message := fmt.Sprintf("Can't Watch Application Health for pod %v(%v): %v",
//			pod.Name, pod.Namespace, err)
//		result := NewCheckResult(m, true, message, nil)
//		klog.Warningf("Premature check failure: %+v", result)
//		metrics.WatchApplicationHealthErrors.Inc()
//		ch.in <- result
//		return nil
//	}
//
//	var appHealthResp *vkvmagentv0.ApplicationHealthResponse
//	recvDone := make(chan bool)
//
//	// receive in a separate goroutine to allow cancellation of this loop while waiting for a message
//	// see https://github.com/grpc/grpc-go/issues/847 for a discussion about why grpc-go doesn't implement a
//	// channel-based interface for streaming receives
//	go func() {
//		for {
//			select {
//			// TODO(guicejg): update to use context-based cancellation via ctx.Done() receive
//			case <-recvDone:
//				klog.InfoS("Streaming Application Health receive cancelled", "pod", klog.KObj(pod))
//				return
//			default:
//				klog.InfoS("Waiting for Application Health status message...",
//					"monitor", m.Name, "watcher", m.watcher, "pod", klog.KObj(pod))
//				appHealthResp, err = stream.Recv()
//
//				if errors.Is(err, io.EOF) {
//					klog.Warning("EOF while receiving application health stream")
//
//					// signal done so the enclosing function exits and gets reconnected to the stream
//					m.watcher.done <- true
//					return
//				}
//
//				if err != nil {
//					message := fmt.Sprintf("Error receiving Application Health stream for pod %v(%v): %v",
//						pod.Name, pod.Namespace, err)
//					result := NewCheckResult(m, true, message, nil)
//					klog.Warningf("Premature check failure: %+v", result)
//					metrics.WatchApplicationHealthStreamErrors.Inc()
//					ch.in <- result
//
//					// signal done so the enclosing function exits and gets reconnected to the stream
//					m.watcher.done <- true
//					return
//				}
//
//				result := NewCheckResult(
//					m, false, "application health stream received status", appHealthResp.PodStatus)
//
//				ch.in <- result
//			}
//		}
//	}()
//
//	// monitor cancellation channel
//	for {
//		select {
//		case <-m.watcher.done:
//			klog.InfoS("Stopping streaming Application Health receiver...", "pod", klog.KObj(pod))
//
//			m.watcher.isWatching = false
//
//			klog.InfoS("App health watcher terminated", "pod", klog.KObj(pod))
//			return nil
//		}
//	}
//
//}

func checkVkvmaConnection(ctx context.Context, m *Monitor) *checkResult {
	return NewCheckResult(m, false, "⚠️  TODO implement this monitor", nil)
}

func checkVkvmaHealth(ctx context.Context, m *Monitor) *checkResult {
	return NewCheckResult(m, false, "⚠️  TODO implement this monitor", nil)
}
