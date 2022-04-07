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

	"github.com/aws/aws-virtual-kubelet/internal/metrics"

	"github.com/aws/aws-virtual-kubelet/internal/config"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/klog/v2"
)

type Subject string

const (
	SubjectUnknown Subject = "unknown"
	SubjectEc2     Subject = "ec2"
	SubjectVkvma   Subject = "vkvma"
	SubjectApp     Subject = "app"
)

type MonitoringState string

const (
	MonitoringStateUnknown   = "unknown"
	MonitoringStateHealthy   = "healthy"
	MonitoringStateUnhealthy = "unhealthy"
)

type Watcher struct {
	isWatching bool
	watch      func(ctx context.Context, monitor *Monitor, ch *CheckHandler) error
	done       chan bool
	results    chan *checkResult
}

type Monitor struct {
	Resource interface{}
	Subject  Subject
	Name     string
	Failures int
	State    MonitoringState
	// "upstream" monitor this monitor's status is reliant upon (if upstream fails, this monitor must also be failing)
	Dependency *Monitor
	check      func(ctx context.Context, monitor *Monitor) *checkResult
	isWatcher  bool
	watcher    *Watcher
}

// checkResult captures the result of a single health check occurrence (and the monitor the check is associated with)
type checkResult struct {
	Monitor   *Monitor
	Failed    bool
	Message   string
	Timestamp time.Time
	Data      interface{} // arbitrary optional data returned by the check
}

func NewCheckResult(monitor *Monitor, failed bool, message string, data interface{}) *checkResult {
	cr := &checkResult{
		Monitor:   monitor,
		Failed:    failed,
		Message:   message,
		Timestamp: time.Now(),
		Data:      data,
	}

	healthConfig := config.Config().HealthConfig

	// creating a check result is generally the only time monitor failures should increment/decrement, so handle it here
	if failed {
		monitor.incrementFailures(healthConfig.UnhealthyThresholdCount)
	} else {
		monitor.resetFailures()
	}

	return cr
}

// NOTE check is just a function _signature_ and m (*Monitor) is set by the time it's actually invoked

func NewMonitor(
	resource interface{}, subject Subject, name string,
	check func(ctx context.Context, m *Monitor) *checkResult) *Monitor {
	m := &Monitor{
		Resource:  resource,
		Subject:   subject,
		Name:      name,
		State:     MonitoringStateUnknown,
		check:     check,
		isWatcher: false,
	}
	klog.InfoS("Created monitor", "monitor", m.Name, "resource", klog.KObj(resource.(*corev1.Pod)))
	return m
}

// NOTE watch is just a function _signature_ and m (*Monitor) is set by the time it's actually invoked

func (m *Monitor) setWatcherFunc(watch func(ctx context.Context, m *Monitor, ch *CheckHandler) error) {
	m.watcher = &Watcher{
		isWatching: false,
		watch:      watch,
	}

	// ensure this monitor isn't also a "Checker"
	m.isWatcher = true
	m.check = nil
	klog.InfoS("Watcher function set", "monitor", m.Name)
}

func (m *Monitor) resetFailures() {
	m.Failures = 0
	klog.V(1).InfoS(
		"Failure counter reset", "monitor", m.Name, "resource", klog.KObj(m.Resource.(*corev1.Pod)))
	m.State = MonitoringStateHealthy
	metrics.HealthCheckStateReset.Inc()
}

func (m *Monitor) incrementFailures(unhealthyThreshold int) {
	m.Failures++
	klog.V(1).InfoS("Incremented failure counter", "monitor", m.Name, "count", m.Failures,
		"resource", klog.KObj(m.Resource.(*corev1.Pod)))

	if m.Failures >= unhealthyThreshold {
		klog.V(1).InfoS(
			"Monitor has reached unhealthy threshold", "monitor", m.Name, "threshold", unhealthyThreshold,
			"resource", klog.KObj(m.Resource.(*corev1.Pod)))
		m.State = MonitoringStateUnhealthy
		metrics.HealthCheckStateUnhealthy.Inc()
	}
}

func (m *Monitor) GetRootFailure() *Monitor {
	klog.V(1).InfoS("Getting root failure", "monitor", m.Name)
	return rootFailure(m)
}

// rootFailure recursively follows the monitor dependency chain to find the "top-most" unhealthy monitor
func rootFailure(m *Monitor) *Monitor {
	parentMonitor := m.Dependency

	// if we are at root of dependency chain, or next upstream monitor is healthy/unknown
	if parentMonitor == nil || parentMonitor.State != MonitoringStateUnhealthy {
		klog.V(1).InfoS("Root failure located", "monitor", m.Name)
		return m
	}

	klog.V(1).InfoS("Traversing parent monitor to locate root failure", "monitor", m.Name, "parent", parentMonitor.Name)
	return rootFailure(parentMonitor)
}

// NOTE if any dependency "up the chain" has failures, don't check the monitor for the current dependency
//  might need a "should check" function...
//  EXCEPTION possibly with watch, since we will receive what we receive regardless....it's an interesting challenge,
//  the bi-directional control of "health"
