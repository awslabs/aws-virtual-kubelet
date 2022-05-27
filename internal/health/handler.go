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
	"sync"

	util "github.com/aws/aws-virtual-kubelet/internal/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type Handler interface {
	receive(ctx context.Context, in chan interface{})
}

type CheckHandler struct {
	// in receives a checkResult to process
	in chan *checkResult
	// IsReceiving is true if handler receiver is currently running
	IsReceiving bool
}

// NewCheckHandler creates a new check handler instance
func NewCheckHandler() *CheckHandler {
	in := make(chan *checkResult)
	ch := &CheckHandler{
		in: in,
	}

	return ch
}

// receive starts a goroutine to receive and process check results
func (ch *CheckHandler) receive(ctx context.Context, wg *sync.WaitGroup) {
	ch.IsReceiving = true

	wg.Add(1)

	// (in a goroutine) receive messages on 'in' and forward to handleCheckResult
	go func() {
		// decrement the WaitGroup counter when the loop exits
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				ch.IsReceiving = false
				klog.InfoS("Handler context cancelled...stopping")
				return
			case result := <-ch.in:
				klog.V(1).InfoS("Received a checkResult to process", "monitor name", result.Monitor.Name,
					"failed?", result.Failed, "message", result.Message)
				ch.handleCheckResult(ctx, result)
			}
		}
	}()
}

// handleCheckResults determines what behavior(s) to perform based on check results received and monitor state
func (ch *CheckHandler) handleCheckResult(ctx context.Context, result *checkResult) {
	klog.V(1).InfoS("Handling checkResult", "checkResult", result)

	monitor := result.Monitor
	pod := monitor.Resource.(*corev1.Pod)

	if result.Failed {
		klog.V(1).InfoS("🟠️️ Check failure", "monitor", monitor.Name, "pod", klog.KObj(pod))
	} else {
		klog.V(1).InfoS("⚪️️ Check success", "monitor", monitor.Name, "pod", klog.KObj(pod))
	}

	// decide how to handle check result
	switch monitor.getState() {
	case MonitoringStateHealthy:
		klog.V(1).InfoS("🟢 Monitor state is HEALTHY", "monitor", monitor.Name, "pod", klog.KObj(pod))

	case MonitoringStateUnhealthy:
		klog.V(1).InfoS("🔴 Monitor state is UNHEALTHY", "monitor", monitor.Name, "pod", klog.KObj(pod))

		switch monitor.Subject {
		case SubjectVkvma:
			klog.InfoS("VKVMA failure...", "monitor", monitor, "pod", klog.KObj(pod))
		case SubjectApp:
			klog.InfoS("App failure...", "monitor", monitor, "pod", klog.KObj(pod))
		default:
			klog.InfoS("Unknown health check subject...ignoring", "monitor", monitor, "pod", klog.KObj(pod))
		}
	}

	if result.Data != nil {
		klog.V(1).InfoS("Processing check data", "pod", klog.KObj(pod))
		if _, ok := result.Data.(*corev1.PodStatus); ok {
			podStatus := result.Data.(*corev1.PodStatus)
			klog.V(1).InfoS("Check data is a `PodStatus`")
			klog.V(1).InfoS("`PodStatus` data contents", "PodStatus", podStatus)

			// ensure IP fields are set to the correct values (agents likely won't/can't set these)
			podStatus.PodIP = pod.Status.PodIP
			podStatus.PodIPs = pod.Status.PodIPs
			podStatus.HostIP = pod.Status.HostIP

			// update pod with combined status
			pod.Status = *podStatus

			notifier := util.GetNotifier()

			// notify pod
			if notifier != nil {
				notifier(pod)
			} else {
				klog.InfoS("⚠️  Unable to notify pod status (handler notifier func not set)", "pod", klog.KObj(pod))
			}
		} else {
			klog.V(1).InfoS("Unknown check data...skipping processing", "data", result.Data)
		}
	}
}
