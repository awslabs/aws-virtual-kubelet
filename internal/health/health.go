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

	health "github.com/aws/aws-virtual-kubelet/proto/grpc/health/v1"

	"github.com/aws/aws-virtual-kubelet/internal/vkvmaclient"
	vkvmagentv0 "github.com/aws/aws-virtual-kubelet/proto/vkvmagent/v0"

	"github.com/aws/aws-virtual-kubelet/internal/config"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/klog/v2"
)

type PodMonitor struct {
	Monitors []*Monitor

	config    config.HealthConfig
	pod       *corev1.Pod
	handler   *CheckHandler
	cancel    context.CancelFunc
	waitGroup *sync.WaitGroup
}

// NewPodMonitor creates monitors appropriate for a pod. The CheckHandler passed in will have its `receive` method
//
//	invoked to receive and process check results, unless a pod specifies a custom handler via pod annotation (in which
//	case _that_ handler's `receive` will be called instead).
func NewPodMonitor(pod *corev1.Pod, handler *CheckHandler) (*PodMonitor, error) {
	cfg := config.Config().HealthConfig
	klog.Infof("Pod Monitor loaded cfg %+v", cfg)

	pm := &PodMonitor{
		config:  cfg,
		pod:     pod,
		handler: setPodHandler(pod, handler),
	}

	pm.createMonitors()

	return pm, nil
}

// setPodHandler allows a pod to specify a different check handler than the default provided
func setPodHandler(pod *corev1.Pod, handler *CheckHandler) *CheckHandler {
	// TODO: check pod annotation for custom check handler and create / set accordingly if present and valid
	return handler
}

// createMonitors creates the monitors associated with a pod.
func (pm *PodMonitor) createMonitors() {
	// create VKVMAgent watcher
	vkvmaWatchMonitor := NewMonitor(pm.pod, SubjectVkvma, "vkvma.watch", nil)
	// connect handler's input channel to monitor
	vkvmaWatchMonitor.handlerReceiver = pm.handler.in
	vkvmaWatchMonitor.isWatcher = true
	vkvmaWatchMonitor.getStream = func(ctx context.Context, m *Monitor) interface{} {
		vc := vkvmaclient.NewVkvmaPodClient(pm.pod)

		hc, err := vc.GetHealthClient(ctx)
		if err != nil {
			klog.ErrorS(err, "Error getting Health client", "monitor", m, "pod", klog.KObj(pm.pod))
			return nil
		}

		stream, err := hc.Watch(ctx, &health.HealthCheckRequest{})
		if err != nil {
			klog.ErrorS(err, "Error calling Watch", "monitor", m, "pod", klog.KObj(pm.pod))
			return nil
		}

		return stream
	}

	// create Application watcher
	appWatchMonitor := NewMonitor(pm.pod, SubjectApp, "app.watch", nil)
	// connect handler's input channel to monitor
	appWatchMonitor.handlerReceiver = pm.handler.in
	appWatchMonitor.isWatcher = true
	appWatchMonitor.getStream = func(ctx context.Context, m *Monitor) interface{} {
		vc := vkvmaclient.NewVkvmaPodClient(pm.pod)

		alc, err := vc.GetApplicationLifecycleClient(ctx)
		if err != nil {
			klog.ErrorS(err, "Error getting Application Lifecycle client", "monitor", m, "pod", klog.KObj(pm.pod))
			return nil
		}

		stream, err := alc.WatchApplicationHealth(ctx, &vkvmagentv0.ApplicationHealthRequest{})
		if err != nil {
			klog.ErrorS(err, "Error calling WatchApplicationHealth", "monitor", m, "pod", klog.KObj(pm.pod))
			return nil
		}

		return stream
	}

	pm.Monitors = []*Monitor{
		vkvmaWatchMonitor,
		appWatchMonitor,
	}
}

// Start activates monitoring
func (pm *PodMonitor) Start(ctx context.Context) {
	klog.InfoS("Starting pod monitor", "pod", klog.KObj(pm.pod))

	// create cancellable context from passed-in context
	pmCtx, cancel := context.WithCancel(ctx)
	// save cancel function reference for later use when stopping monitors/handler
	pm.cancel = cancel

	// create wait group to wait for monitor-related goroutine to stop
	pm.waitGroup = &sync.WaitGroup{}

	// start check handler's receiver
	pm.handler.receive(pmCtx, pm.waitGroup)

	// start monitors
	for _, m := range pm.Monitors {
		m.Run(pmCtx, pm.waitGroup)
	}
}

// Stop deactivates monitoring
func (pm *PodMonitor) Stop() {
	klog.InfoS("Stopping pod monitor", "pod", klog.KObj(pm.pod))

	// cancel the monitor(s)
	pm.cancel()

	// TODO(guicejg): add a timeout to this wait? (if the goroutines don't exit from cancel() we can't cancel them
	//  any other way, but we could at least WARN and/or emit a metric)
	// wait for all goroutines to exit
	pm.waitGroup.Wait()

	klog.InfoS("All monitors cancelled", "pod", klog.KObj(pm.pod))
}
