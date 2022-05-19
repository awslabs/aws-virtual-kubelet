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
	"sync"
	"time"

	health "github.com/aws/aws-virtual-kubelet/proto/grpc/health/v1"

	vkvmagentv0 "github.com/aws/aws-virtual-kubelet/proto/vkvmagent/v0"

	"github.com/aws/aws-virtual-kubelet/internal/metrics"

	"github.com/aws/aws-virtual-kubelet/internal/config"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/klog/v2"
)

// Subject is the specific target of a monitoring check (usually a component of a resource)
type Subject string

const (
	// SubjectUnknown is a subject without explicit handling
	SubjectUnknown Subject = "unknown"
	// SubjectVkvma is the Virtual Kubelet Virtual Machine Agent (the gRPC server itself)
	SubjectVkvma Subject = "vkvma"
	// SubjectApp is the ApplicationLifecycle service of the VKVMAgent
	SubjectApp Subject = "app"
)

// MonitoringState represents the state of the resource being monitored
type MonitoringState string

const (
	// MonitoringStateUnknown represents the state where we don't know if the resource is healthy or not
	MonitoringStateUnknown = "unknown"
	// MonitoringStateHealthy represents a healthy resource
	MonitoringStateHealthy = "healthy"
	// MonitoringStateUnhealthy represents an unhealth resource
	MonitoringStateUnhealthy = "unhealthy"
)

//// TODO(guicejg): document Watcher
//type Watcher struct {
//	isWatching bool
//	watch      func(ctx context.Context, monitor *Monitor, ch *CheckHandler) error
//	done       chan bool
//	results    chan *checkResult
//}

// Monitor is the set of properties associate with a monitor instance
type Monitor struct {
	// Resource being monitored (can be of any type)
	Resource interface{}
	// Subject type of the monitored resource
	Subject Subject
	// Name of the monitored resource
	Name string
	// Failures count of the monitored resource (check and watch type monitors both accumulate here)
	Failures int
	// State of the monitored resource
	State MonitoringState
	// IsMonitoring is true if monitoring is currently running
	IsMonitoring bool
	// check function to run to check resource and obtain a CheckResult from (for check type monitors)
	check func(ctx context.Context, monitor *Monitor) *checkResult
	// isWatcher is true if this is a watch type monitor (TODO(guicejg): could we just check if stream func is null?)
	isWatcher bool
	// stream function that returns a stream to receive from
	getStream func(ctx context.Context, monitor *Monitor) interface{}
	// handlerReceiver is the channel that the check handler receives check results on
	handlerReceiver chan *checkResult

	// sync.RWMutex enables synchronization when a monitor's properties are potentially updated in multiple goroutines
	sync.RWMutex
}

// checkResult captures the result of a single health check occurrence (and the monitor the check is associated with)
type checkResult struct {
	// Monitor is the monitor associated with the check result
	Monitor *Monitor
	// Failed is true if the check failed
	Failed bool
	// Message contains additional details about the check result
	Message string
	// Timestamp is the point in time when the result was obtained
	Timestamp time.Time
	// Data is a container for arbitrary (optional) data that may be returned with a check result (e.g. a PodStatus)
	Data interface{}
}

// NewCheckResult creates a new check result for a particular monitor, failure state, message and (optional) data.
//  This is also where the monitor state is updated since the state should only change based on the result of a check.
func NewCheckResult(monitor *Monitor, failed bool, message string, data interface{}) *checkResult {
	cr := &checkResult{
		Monitor:   monitor,
		Failed:    failed,
		Message:   message,
		Timestamp: time.Now(),
		Data:      data,
	}

	healthConfig := config.Config().HealthConfig

	// increment or reset failure counters
	if failed {
		monitor.incrementFailures(healthConfig.UnhealthyThresholdCount)
	} else {
		monitor.resetFailures()
	}

	return cr
}

// NOTE check is just a function _signature_ and m (*Monitor) is set by the time it's actually invoked (so the apparent
//  or potential recursion is safe due to enforced runtime ordering)

// NewMonitor creates a new Monitor for a resource, targeting a subject (component) of that resource.  Each individual
//  monitor is associated with a check function that knows how to validate the state of the subject.  The function
//  must return a checkResult in _every case_ (no error-handling is applied to check function return values).
//
//  Check functions are executed every `HealthCheckIntervalSeconds` seconds and the result is sent to a channel provided
//  at Monitor instantiation.  This `handlerReceiver` channel should be monitored by an upstream handler function that
//  can decide  what actions to take based on check results.
//
//  Stopping a monitor is accomplished by calling `cancel()` on the context provided at instantiation.  This context
//  must be a context.WithCancel and is owned by the creator of the monitor (e.g. PodMonitor).  A WaitGroup parameter
//  is also present, which allows tracking of goroutines created by a monitor.  This helps ensure goroutines do not leak
//  over time.
func NewMonitor(
	resource interface{}, subject Subject, name string,
	check func(ctx context.Context, m *Monitor) *checkResult) *Monitor {
	m := &Monitor{
		Resource:     resource,
		Subject:      subject,
		Name:         name,
		State:        MonitoringStateUnknown,
		IsMonitoring: false,
		check:        check,
		isWatcher:    false,
	}
	klog.InfoS("Created monitor", "monitor", m)
	return m
}

// Run executes the monitoring loop for a monitor
func (m *Monitor) Run(ctx context.Context, wg *sync.WaitGroup) {
	cfg := config.Config()

	m.IsMonitoring = true

	if m.isWatcher {
		m.startWatchLoop(ctx, wg, *cfg)
	} else {
		m.startCheckLoop(ctx, wg, *cfg)
	}

	// TODO(guicejg): can we convince klog to use a property accessor that is wrapped in a RLock() / RUnlock()?
	// avoid DATA RACE in klog property read below
	m.RLock()
	defer m.RUnlock()

	klog.InfoS("Monitor started", "monitor", m)
}

func (m *Monitor) startWatchLoop(ctx context.Context, wg *sync.WaitGroup, cfg config.ProviderConfig) {
	klog.V(1).InfoS("Starting watch loop", "monitor", m)

	wg.Add(1)

	// TODO(guicejg): put this in a conditional to allow handling of non-Pods
	pod := m.Resource.(*corev1.Pod)

	go func() {
		// decrement the WaitGroup counter when the loop exits
		defer wg.Done()

		var stream interface{}

		for {
			// (re)obtain stream if needed
			if stream == nil {
				klog.InfoS("Connecting stream...", "monitor", m, "pod", klog.KObj(pod.DeepCopy()))
				stream = m.getStream(ctx, m)
			}

			switch typedStream := stream.(type) {
			case health.Health_WatchClient:
				klog.V(1).InfoS("Stream is from health.Health_WatchClient")

				var healthResp *health.HealthCheckResponse
				var err error

				select {
				// cancellation requested via context
				case <-ctx.Done():
					klog.InfoS("Monitor stopping...", "monitor", m)
					m.Lock()
					m.State = MonitoringStateUnknown
					m.IsMonitoring = false
					m.Unlock()
					return
				// setup complete, wait for streaming result
				default:
					klog.V(1).InfoS("Waiting for VKVMAgent Health status message...",
						"monitor", m.Name, "pod", klog.KObj(pod))
					healthResp, err = typedStream.Recv()

					// Error from app health stream, generate failed check result and attempt to reconnect
					if err != nil {
						// handle case where error is due to context cancellation
						select {
						case <-ctx.Done():
							klog.InfoS("Monitor stopping...", "monitor", m)
							m.Lock()
							m.State = MonitoringStateUnknown
							m.IsMonitoring = false
							m.Unlock()
							return
						default:
							message := fmt.Sprintf("Error receiving VKVMAgent Health stream for pod %v(%v): %v",
								pod.Name, pod.Namespace, err)

							result := NewCheckResult(m, true, message, healthResp)

							klog.Warningf("Premature check failure: %+v", result)

							metrics.WatchApplicationHealthErrors.Inc() // TODO(guicejg): this is the wrong metric...

							m.handlerReceiver <- result

							time.Sleep(time.Duration(cfg.HealthConfig.StreamRetryIntervalSeconds) * time.Second)

							// force gRPC to reconnect stream
							stream = nil
							continue
						}
					}

					// received a result from vkvma health stream, send to handler and loop to wait for the next one
					result := NewCheckResult(
						m, false, "application health stream received status", nil)

					m.handlerReceiver <- result
				}
			case vkvmagentv0.ApplicationLifecycle_WatchApplicationHealthClient:
				klog.V(1).InfoS("Stream is from ApplicationLifecycle_WatchApplicationHealthClient")

				var appHealthResp *vkvmagentv0.ApplicationHealthResponse
				var err error

				select {
				// cancellation requested via context
				case <-ctx.Done():
					klog.InfoS("Monitor stopping...", "monitor", m)
					m.Lock()
					m.State = MonitoringStateUnknown
					m.IsMonitoring = false
					m.Unlock()
					return
				// setup complete, block for streaming result
				default:
					klog.V(1).InfoS("Waiting for Application Health status message...",
						"monitor", m.Name, "pod", klog.KObj(pod))
					appHealthResp, err = typedStream.Recv()

					// Error from app health stream, generate failed check result and attempt to reconnect
					if err != nil {
						select {
						case <-ctx.Done():
							klog.InfoS("Monitor stopping...", "monitor", m)
							m.Lock()
							m.State = MonitoringStateUnknown
							m.IsMonitoring = false
							m.Unlock()
							return
						default:
							message := fmt.Sprintf("Error receiving Application Health stream for pod %v(%v): %v",
								pod.Name, pod.Namespace, err)

							result := NewCheckResult(m, true, message, nil)

							klog.Warningf("Premature check failure: %+v", result)

							metrics.WatchApplicationHealthStreamErrors.Inc()

							m.handlerReceiver <- result

							time.Sleep(time.Duration(cfg.HealthConfig.StreamRetryIntervalSeconds) * time.Second)

							// force gRPC to reconnect stream
							stream = nil
							continue
						}
					}

					// received a result from app health stream, send to handler and loop to wait for the next one
					result := NewCheckResult(
						m, false, "application health stream received status", appHealthResp.PodStatus)

					m.handlerReceiver <- result
				}
			default:
				// still listen for context cancellation even when we can't identify the stream type
				// TODO(guicejg): how can we not duplicate this code in every watch type handler?
				select {
				// cancellation requested via context
				case <-ctx.Done():
					klog.InfoS("Monitor stopping...", "monitor", m)
					m.Lock()
					m.State = MonitoringStateUnknown
					m.IsMonitoring = false
					m.Unlock()
					return
				default:
					if stream == nil {
						result := NewCheckResult(m, true, "Health check stream is nil", nil)
						m.handlerReceiver <- result
						time.Sleep(time.Duration(cfg.HealthConfig.StreamRetryIntervalSeconds) * time.Second)
					} else {
						klog.InfoS("Stream type is unknown, ignoring (⚠️  this monitor will do nothing)")
						time.Sleep(time.Duration(cfg.HealthConfig.HealthCheckIntervalSeconds) * time.Second)
					}
				}
			}
		}
	}()
}

func (m *Monitor) startCheckLoop(ctx context.Context, wg *sync.WaitGroup, cfg config.ProviderConfig) {
	wg.Add(1)

	go func() {
		// decrement the WaitGroup counter when the loop exits
		defer wg.Done()

		for {
			// NOTE checks are responsible for all error handling (the only result of a check execution should be a
			//  check result)

			klog.InfoS("Initiating check", "monitor", m)
			result := m.check(ctx, m)

			select {
			// cancellation requested via context
			case <-ctx.Done():
				klog.InfoS("Monitor stopping...", "monitor", m)
				m.Lock()
				m.State = MonitoringStateUnknown
				m.IsMonitoring = false
				m.Unlock()
				return
			// handler receiver ready to receive a result
			case m.handlerReceiver <- result:
				klog.InfoS("Sending check result to handler receiver", "pod", klog.KObj(m.Resource.(*corev1.Pod)),
					"monitor", m, "result", result)
			}

			klog.InfoS("Sleeping until next Check Interval", "pod", klog.KObj(m.Resource.(*corev1.Pod)),
				"interval (seconds)", cfg.HealthConfig.HealthCheckIntervalSeconds)
			time.Sleep(time.Duration(cfg.HealthConfig.HealthCheckIntervalSeconds))
		}
	}()
}

func (m *Monitor) resetFailures() {
	m.Lock()
	defer m.Unlock()

	m.Failures = 0

	klog.V(1).InfoS(
		"Failure counter reset", "monitor", m.Name, "resource", klog.KObj(m.Resource.(*corev1.Pod)))

	m.State = MonitoringStateHealthy

	metrics.HealthCheckStateReset.Inc()
}

func (m *Monitor) incrementFailures(unhealthyThreshold int) {
	m.Lock()
	defer m.Unlock()

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

// getState gets the current monitoring state (concurrency safe)
func (m *Monitor) getState() MonitoringState {
	m.RLock()
	state := m.State
	m.RUnlock()

	return state
}

// TODO(guicejg): add unit test(s) for this
func (m *Monitor) String() string {
	// TODO(guicejg): check verbosity level and log additional info (V(2) should log entire structs e.g. "%+v",m)

	mString := m.Name + ":"

	// TODO(guicejg): consolidate locking into monitor.GetState function
	// monitor state read requires locking
	m.RLock()
	defer m.RUnlock()

	switch m.State {
	case MonitoringStateHealthy:
		if m.Failures > 0 {
			mString += "🟠"
		} else {
			mString += "🟢"
		}
	case MonitoringStateUnhealthy:
		mString += "🔴"
	case MonitoringStateUnknown:
		mString += "🔵"
	}

	return mString
}
