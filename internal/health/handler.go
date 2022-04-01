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

	util "github.com/aws/aws-virtual-kubelet/internal/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type Handler interface {
	receive(ctx context.Context, in chan interface{})
}

type CheckHandler struct {
	in   chan *checkResult
	done chan bool
}

func NewCheckHandler(notifier func(*corev1.Pod)) *CheckHandler {
	in := make(chan *checkResult)
	done := make(chan bool)
	ch := &CheckHandler{
		in:   in,
		done: done,
	}

	return ch
}

func (ch *CheckHandler) receive(ctx context.Context) {
	// (in a goroutine) receive messages on in and forward to handleCheckResult
	go func() {
		for {
			select {
			case result := <-ch.in:
				klog.InfoS("Received a checkResult to process", "monitor name", result.Monitor.Name,
					"failed?", result.Failed, "message", result.Message)
				ch.handleCheckResult(ctx, result)
			case <-ch.done:
				klog.InfoS("Handler received 'done'...stopping")
				return
			}
		}
	}()
}

func (ch *CheckHandler) handleCheckResult(ctx context.Context, result *checkResult) {
	klog.InfoS("Handling checkResult", "checkResult", result)

	monitor := result.Monitor
	pod := monitor.Resource.(*corev1.Pod)

	if result.Failed {
		klog.InfoS("🟠️️ Check failure", "monitor", monitor.Name, "pod", klog.KObj(pod))
	} else {
		klog.InfoS("⚪️️ Check success", "monitor", monitor.Name, "pod", klog.KObj(pod))
	}

	// decide how to handle check result
	switch monitor.State {
	case MonitoringStateHealthy:
		klog.InfoS("🟢 Monitor state is HEALTHY", "monitor", monitor.Name, "pod", klog.KObj(pod))

	case MonitoringStateUnhealthy:
		klog.InfoS("🔴 Monitor state is UNHEALTHY", "monitor", monitor.Name, "pod", klog.KObj(pod))
		//get most upstream failing monitor and take steps based on it (may return the same failing monitor we have)
		primaryFailure := monitor.GetRootFailure()

		switch primaryFailure.Subject {
		case SubjectEc2:
			klog.InfoS("EC2 failure...", "pod", klog.KObj(pod))

			// TODO add metrics here

			err := util.ReplaceCompute(ctx, pod)
			if err != nil {
				panic("retry here instead")
			}
		case SubjectVkvma:
			klog.InfoS("VKVMA failure...", "pod", klog.KObj(pod))
			//p.deletePodSkipApp(ctx, pod)
			//_ = p.CreatePod(ctx, pod)
		case SubjectApp:
			klog.InfoS("App failure...ignoring", "pod", klog.KObj(pod))
			//p.deletePodSkipApp(ctx, pod)
			//_ = p.CreatePod(ctx, pod)
		default:
			klog.InfoS("Unknown health check subject...ignoring", "pod", klog.KObj(pod))
		}

	case MonitoringStateUnknown:
		klog.InfoS("🔵 Monitor state is UNKNOWN", "monitor", monitor.Name, "pod", klog.KObj(pod))
		//// get most upstream failing monitor and take steps based on it (may return the same failing monitor we have)
		//rootFailure := monitor.GetRootFailure()
		//
		//switch rootFailure.Subject {
		//case SubjectEc2:
		//	// treat unknown EC2 failure as fatal and recreate (skipping any app communication as it's likely down)
		//	p.deletePodSkipApp(ctx, pod)
		//	_ = p.CreatePod(ctx, pod)
		//case SubjectVkvma:
		//	klog.InfoS("VKVMA state unknown...", "pod", klog.KObj(pod))
		////	p.deletePodSkipApp(ctx, pod)
		////	_ = p.CreatePod(ctx, pod)
		//case SubjectApp:
		//	klog.InfoS("App state unknown...", "pod", klog.KObj(pod))
		////	p.deletePodSkipApp(ctx, pod)
		////	_ = p.CreatePod(ctx, pod)
		//default:
		//	klog.InfoS("Unknown health check subject...ignoring", "pod", klog.KObj(pod))
		//}
	}

	if result.Data != nil {
		klog.InfoS("Processing check data", "pod", klog.KObj(pod))
		if _, ok := result.Data.(*corev1.PodStatus); ok {
			podStatus := result.Data.(*corev1.PodStatus)
			klog.InfoS("Check data is a `PodStatus`")
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
			klog.InfoS("Unknown check data...skipping processing", "data", result.Data)
		}
	}
}
