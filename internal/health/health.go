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
	"time"

	"github.com/aws/aws-virtual-kubelet/internal/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type PodMonitor struct {
	config       config.HealthConfig
	pod          *corev1.Pod
	handler      *CheckHandler
	monitors     []*Monitor
	isMonitoring bool
	done         chan bool
}

func NewPodMonitor(pod *corev1.Pod, handler *CheckHandler) (*PodMonitor, error) {

	cfg := config.Config().HealthConfig
	klog.Infof("Pod Monitor loaded cfg %+v", cfg)

	pm := &PodMonitor{
		pod:     pod,
		config:  cfg,
		handler: handler,
	}

	pm.configureMonitors()

	return pm, nil
}

func (pm *PodMonitor) configureMonitors() {
	// create monitors and dependencies (see docs/health/monitor.puml)

	ec2StatusMonitor := NewMonitor(pm.pod, SubjectEc2, "ec2.status", checkEc2Status)

	//vkvmaConnectMonitor := NewMonitor(pm.pod, SubjectVkvma, "vkvma.connect", checkVkvmaConnection)
	//vkvmaConnectMonitor.Dependency = ec2StatusMonitor
	//
	//vkvmaHealthMonitor := NewMonitor(pm.pod, SubjectVkvma, "vkvma.health", checkVkvmaHealth)
	//vkvmaHealthMonitor.Dependency = vkvmaConnectMonitor
	//
	//appHealthMonitor := NewMonitor(pm.pod, SubjectApp, "app.health", checkAppHealth)
	//appHealthMonitor.Dependency = vkvmaHealthMonitor

	//create a (streaming) "watcher" monitor instead of a standard (polling) "checker" monitor like the above
	appWatchMonitor := NewMonitor(pm.pod, SubjectApp, "app.monitor", nil)
	appWatchMonitor.setWatcherFunc(watchAppHealth)
	//appWatchMonitor.Dependency = vkvmaConnectMonitor
	appWatchMonitor.Dependency = ec2StatusMonitor // connect directly to ec2 until intermediate monitors are re-enabled

	pm.monitors = []*Monitor{
		ec2StatusMonitor,
		//vkvmaConnectMonitor,
		//vkvmaHealthMonitor,
		//appHealthMonitor,
		appWatchMonitor,
	}
}

func (pm *PodMonitor) Start(ctx context.Context) {
	klog.Infof("Start pod monitor for pod %v(%v)", pm.pod.Name, pm.pod.Namespace)

	done := make(chan bool)
	pm.done = done

	go pm.monitorLoop(ctx, pm.done)
	go pm.handler.receive(ctx)
}

func (pm *PodMonitor) monitorLoop(ctx context.Context, done chan bool) {
	for {
		klog.V(1).InfoS("Start of monitor loop", "pod", klog.KObj(pm.pod))
		select {
		case <-done:
			klog.InfoS("Monitor loop terminated", "pod", klog.KObj(pm.pod))
			pm.isMonitoring = false
			klog.InfoS("Stopping handler receiver...", "pod", klog.KObj(pm.pod))
			pm.handler.done <- true
			return
		default:
			pm.isMonitoring = true

			klog.Infof("Checking %d monitors for pod %v(%v)",
				len(pm.monitors), pm.pod.Name, pm.pod.Namespace)

			var check *checkResult
			var m *Monitor
			var err error
			for _, m = range pm.monitors {
				if m.isWatcher {
					klog.V(1).InfoS(
						"Ensuring watch type monitor is running", "monitor", m.Name, "pod", klog.KObj(pm.pod))
					// pm.background is a noop if the watcher has already been started
					pm.background(ctx, m)
				} else {
					klog.InfoS(
						"Running check type monitor check func", "monitor", m.Name, "pod", klog.KObj(pm.pod))
					check = m.check(ctx, m)
					if err != nil {
						klog.Warningf("Premature failure checking monitor %+v: %v", m, err)
					}

					klog.InfoS(
						"Monitor check completed...sending result to handler", "monitor", m.Name, "state", m.State,
						"failed?", check.Failed, "resource", klog.KObj(m.Resource.(*corev1.Pod)))

					pm.handler.in <- check
				}
			}

			time.Sleep(time.Duration(pm.config.HealthCheckIntervalSeconds) * time.Second)
		}
	}
}

func (pm *PodMonitor) Stop(ctx context.Context) {
	klog.InfoS("Stopping pod monitor", "pod", klog.KObj(pm.pod))

	klog.V(1).InfoS("Pod monitoring state", "monitoring", pm.isMonitoring)
	if pm.isMonitoring {
		var m *Monitor

		// stop all monitors and mark state as Unknown
		for _, m = range pm.monitors {
			if m.isWatcher {
				klog.V(1).InfoS("Stopping WATCH type monitor", "monitor",
					m.Name, "pod", klog.KObj(pm.pod))
				m.watcher.done <- true
				m.watcher.isWatching = false
				klog.V(1).InfoS("Sent DONE to watcher channel",
					"monitor", m.Name, "watcher", m.watcher)
			} else {
				klog.V(1).InfoS("Stopping CHECK type monitor", "monitor",
					m.Name, "pod", klog.KObj(pm.pod))
			}

			m.resetFailures()
			m.State = MonitoringStateUnknown
		}

		// stop monitoring main loop
		klog.InfoS("Sending DONE to pod monitoring goroutine", "pod", klog.KObj(pm.pod))
		pm.done <- true
		klog.InfoS("Sent DONE to pod monitoring goroutine", "pod", klog.KObj(pm.pod))
	}
}

func (pm *PodMonitor) background(ctx context.Context, m *Monitor) {

	if m.watcher.isWatching {
		klog.V(1).InfoS("Ignoring request to start watch-based monitor goroutine (already watching)")
		return // already watching in the background
	} else {
		klog.V(1).InfoS("Starting watch-based monitor goroutine")
	}

	m.watcher.done = make(chan bool)

	go func() {
		for {
			// a watch type monitor function starts in a goroutine that checks `done` to know when to stop working and
			//  sends results directly to the check handler
			err := m.watcher.watch(ctx, m, pm.handler)
			if err != nil {
				klog.ErrorS(err, "Can't start watcher", "monitor", m, "pod", klog.KObj(pm.pod))
			}

			klog.V(1).InfoS("Watcher function returned...awaiting restart", "pod", klog.KObj(m.Resource.(*corev1.Pod)))
		}
	}()
}
