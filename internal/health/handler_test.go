package health

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	util "github.com/aws/aws-virtual-kubelet/internal/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-virtual-kubelet/internal/config"

	v1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
)

func TestNewCheckHandler(t *testing.T) {
	tests := []struct {
		name string
		want *CheckHandler
	}{
		{
			name: "Creates a new check handler correctly",
			want: &CheckHandler{
				in: make(chan *checkResult),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCheckHandler()
			if !assert.IsType(t, tt.want.in, got.in, fmt.Sprintf("NewCheckHandler() = %v, want %v", got, tt.want)) {
				t.Errorf("NewCheckHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReceive(t *testing.T) {
	handler := NewCheckHandler()

	ctx, cancel := context.WithCancel(context.Background())

	// create wait group to wait for monitor goroutine to stop
	waitGroup := sync.WaitGroup{}

	// start check handler receiver (with wait group)
	handler.receive(ctx, &waitGroup)

	// simulate some other activity
	time.Sleep(time.Second * 1)

	// cancel the receiver
	cancel()

	// wait for and assert that the receiver goroutine ended
	waitGroup.Wait()
	assert.False(t, handler.IsReceiving, "Receiving should be stopped.")
}

func TestReceiveWithCheckResults(t *testing.T) {
	// create minimum config
	_ = config.InitConfig(&config.DirectLoader{DirectConfig: config.ProviderConfig{
		HealthConfig: config.HealthConfig{
			HealthCheckIntervalSeconds: 1,
			UnhealthyThresholdCount:    1,
		},
	}})

	handler := NewCheckHandler()

	ctx, cancel := context.WithCancel(context.Background())

	// create wait group to wait for receiver goroutine to stop
	waitGroup := sync.WaitGroup{}

	// start check handler receiver (with wait group)
	handler.receive(ctx, &waitGroup)

	// create fake pod
	fakePod := &v1.Pod{}

	// create monitor for fake pod
	m := NewMonitor(fakePod, SubjectVkvma, "TestMonitor", nil)

	// generate a successful check result
	handler.in <- NewCheckResult(m, false, "Successful check result", nil)

	// check that monitoring state is now healthy
	assert.Equal(t, MonitoringStateHealthy, string(m.State))

	// sleep to allow the assertion above to register before the monitor fails due to the next result
	time.Sleep(time.Second * 1)

	// generate a failed check result
	handler.in <- NewCheckResult(m, true, "Failing check result", nil)

	// check that monitoring state is now healthy (unhealthy check threshold is 1 for this test)
	assert.Equal(t, MonitoringStateUnhealthy, string(m.State))

	// cancel the receiver
	cancel()

	// wait for and assert that the receiver goroutine ended
	waitGroup.Wait()
	assert.False(t, handler.IsReceiving, "Receiving should be stopped.")
}

func TestReceiveSubjects(t *testing.T) {
	// create minimum config
	_ = config.InitConfig(&config.DirectLoader{DirectConfig: config.ProviderConfig{
		HealthConfig: config.HealthConfig{
			HealthCheckIntervalSeconds: 1,
			UnhealthyThresholdCount:    1,
		},
	}})

	handler := NewCheckHandler()

	ctx, cancel := context.WithCancel(context.Background())

	// create wait group to wait for receiver goroutine to stop
	waitGroup := sync.WaitGroup{}

	// start check handler receiver (with wait group)
	handler.receive(ctx, &waitGroup)

	// create fake pod
	fakePod := &v1.Pod{}

	// create monitor for fake pod (app)
	m := NewMonitor(fakePod, SubjectApp, "TestMonitor", nil)

	// generate a failed check result for the App subject
	handler.in <- NewCheckResult(m, true, "Failing check result (app)", nil)

	// check that monitoring state is now healthy (unhealthy check threshold is 1 for this test)
	assert.Equal(t, MonitoringStateUnhealthy, string(m.State))
	assert.Equal(t, SubjectApp, m.Subject)

	// create monitor for fake pod (unknown)
	m = NewMonitor(fakePod, SubjectUnknown, "TestMonitor", nil)

	// generate a failed check result for an Unknown subject
	handler.in <- NewCheckResult(m, true, "Failing check result (unknown)", nil)

	// check that monitoring state is now healthy (unhealthy check threshold is 1 for this test)
	assert.Equal(t, MonitoringStateUnhealthy, string(m.State))
	assert.Equal(t, SubjectUnknown, m.Subject)

	// cancel the receiver
	cancel()

	// wait for and assert that the receiver goroutine ended
	waitGroup.Wait()
	assert.False(t, handler.IsReceiving, "Receiving should be stopped.")
}

func Test_checkHandlerWithData(t *testing.T) {
	// create minimum config
	_ = config.InitConfig(&config.DirectLoader{DirectConfig: config.ProviderConfig{
		HealthConfig: config.HealthConfig{
			HealthCheckIntervalSeconds: 1,
			UnhealthyThresholdCount:    1,
		},
	}})

	handler := NewCheckHandler()

	// create fake pod with data
	fakePod := &v1.Pod{TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1.PodSpec{},
		Status: v1.PodStatus{
			Phase:      v1.PodUnknown,
			Conditions: nil,
			Message:    "TestMessage",
			HostIP:     "127.0.0.1",
			PodIP:      "10.0.0.0",
			PodIPs: []v1.PodIP{{
				IP: "0.0.0.0",
			}},
		},
	}

	// create monitor for fake pod
	m := NewMonitor(fakePod, SubjectVkvma, "TestMonitor", nil)

	expectedPodStatus := &v1.PodStatus{
		Phase: v1.PodRunning,
		Conditions: []v1.PodCondition{{
			Type:    v1.ContainersReady,
			Status:  v1.ConditionTrue,
			Message: "Status set by VKVMAgent",
		}},
		Message: "TestMessage",
		HostIP:  "replace-me",
		PodIP:   "replace-me",
		PodIPs:  nil, // also replace-me
	}

	// generate a successful check result with status data
	handler.handleCheckResult(context.TODO(), NewCheckResult(m, false, "Successful check result", expectedPodStatus))

	// verify pod status elements were updated appropriately
	assert.Equal(t, expectedPodStatus.Phase, fakePod.Status.Phase)
	assert.Equal(t, expectedPodStatus.Conditions, fakePod.Status.Conditions)
	assert.Equal(t, expectedPodStatus.Message, fakePod.Status.Message)
	assert.Equal(t, "127.0.0.1", fakePod.Status.HostIP)
	assert.Equal(t, "10.0.0.0", fakePod.Status.PodIP)
	assert.Equal(t, []v1.PodIP{{IP: "0.0.0.0"}}, fakePod.Status.PodIPs)
}

func Test_checkHandlerWithFailingData(t *testing.T) {
	// create minimum config
	_ = config.InitConfig(&config.DirectLoader{DirectConfig: config.ProviderConfig{
		HealthConfig: config.HealthConfig{
			HealthCheckIntervalSeconds: 1,
			UnhealthyThresholdCount:    1,
		},
	}})

	handler := NewCheckHandler()

	// create fake pod with data
	fakePod := &v1.Pod{TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1.PodSpec{},
		Status: v1.PodStatus{
			Phase:      v1.PodUnknown,
			Conditions: nil,
			Message:    "TestMessage",
			HostIP:     "127.0.0.1",
			PodIP:      "10.0.0.0",
			PodIPs: []v1.PodIP{{
				IP: "0.0.0.0",
			}},
		},
	}

	// create monitor for fake pod
	m := NewMonitor(fakePod, SubjectVkvma, "TestMonitor", nil)

	expectedPodStatus := &v1.PodStatus{
		Phase: v1.PodFailed,
		Conditions: []v1.PodCondition{{
			Type:    v1.PodReady,
			Status:  v1.ConditionFalse,
			Message: "Failing Status set by VKVMAgent",
		}},
		Message: "TestMessage2",
		HostIP:  "replace-me-again",
		PodIP:   "replace-me-again",
		PodIPs:  nil, // also replace-me-again
	}

	// generate a successful check result with status data
	handler.handleCheckResult(context.TODO(), NewCheckResult(m, false, "Successful check result", expectedPodStatus))

	// verify pod status elements were updated appropriately
	assert.Equal(t, expectedPodStatus.Phase, fakePod.Status.Phase)
	assert.Equal(t, expectedPodStatus.Conditions, fakePod.Status.Conditions)
	assert.Equal(t, expectedPodStatus.Message, fakePod.Status.Message)
	assert.Equal(t, "127.0.0.1", fakePod.Status.HostIP)
	assert.Equal(t, "10.0.0.0", fakePod.Status.PodIP)
	assert.Equal(t, []v1.PodIP{{IP: "0.0.0.0"}}, fakePod.Status.PodIPs)
}

func Test_checkHandlerWithNotifier(t *testing.T) {
	// create minimum config
	_ = config.InitConfig(&config.DirectLoader{DirectConfig: config.ProviderConfig{
		HealthConfig: config.HealthConfig{
			HealthCheckIntervalSeconds: 1,
			UnhealthyThresholdCount:    1,
		},
	}})

	util.SetNotifier(func(pod *v1.Pod) {
		log.Printf("Notifier called with pod %+v", pod)
	})

	handler := NewCheckHandler()

	// create fake pod with data
	fakePod := &v1.Pod{TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1.PodSpec{},
		Status: v1.PodStatus{
			Phase:      v1.PodUnknown,
			Conditions: nil,
			Message:    "TestMessage",
			HostIP:     "127.0.0.1",
			PodIP:      "10.0.0.0",
			PodIPs: []v1.PodIP{{
				IP: "0.0.0.0",
			}},
		},
	}

	// create monitor for fake pod
	m := NewMonitor(fakePod, SubjectVkvma, "TestMonitor", nil)

	expectedPodStatus := &v1.PodStatus{
		Phase: v1.PodFailed,
		Conditions: []v1.PodCondition{{
			Type:    v1.PodReady,
			Status:  v1.ConditionFalse,
			Message: "Failing Status set by VKVMAgent",
		}},
		Message: "TestMessage2",
		HostIP:  "replace-me-again",
		PodIP:   "replace-me-again",
		PodIPs:  nil, // also replace-me-again
	}

	// generate a successful check result with status data
	handler.handleCheckResult(context.TODO(), NewCheckResult(m, false, "Successful check result", expectedPodStatus))

	// verify pod status elements were updated appropriately
	assert.Equal(t, expectedPodStatus.Phase, fakePod.Status.Phase)
	assert.Equal(t, expectedPodStatus.Conditions, fakePod.Status.Conditions)
	assert.Equal(t, expectedPodStatus.Message, fakePod.Status.Message)
	assert.Equal(t, "127.0.0.1", fakePod.Status.HostIP)
	assert.Equal(t, "10.0.0.0", fakePod.Status.PodIP)
	assert.Equal(t, []v1.PodIP{{IP: "0.0.0.0"}}, fakePod.Status.PodIPs)
}

func Test_checkHandlerWithUnknownData(t *testing.T) {
	// create minimum config
	_ = config.InitConfig(&config.DirectLoader{DirectConfig: config.ProviderConfig{
		HealthConfig: config.HealthConfig{
			HealthCheckIntervalSeconds: 1,
			UnhealthyThresholdCount:    1,
		},
	}})

	handler := NewCheckHandler()

	// create fake pod with data
	fakePod := &v1.Pod{TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1.PodSpec{},
		Status: v1.PodStatus{
			Phase:      v1.PodUnknown,
			Conditions: nil,
			Message:    "TestMessage",
			HostIP:     "127.0.0.1",
			PodIP:      "10.0.0.0",
			PodIPs: []v1.PodIP{{
				IP: "0.0.0.0",
			}},
		},
	}

	// create monitor for fake pod
	m := NewMonitor(fakePod, SubjectVkvma, "TestMonitor", nil)

	unknownData := &v1.Namespace{}

	// generate a successful check result with status data
	handler.handleCheckResult(context.TODO(), NewCheckResult(m, false, "Successful check result", unknownData))

	// verify pod status did not change
	assert.Equal(t, v1.PodUnknown, fakePod.Status.Phase)
	assert.Equal(t, []v1.PodCondition(nil), fakePod.Status.Conditions)
	assert.Equal(t, "TestMessage", fakePod.Status.Message)
	assert.Equal(t, "127.0.0.1", fakePod.Status.HostIP)
	assert.Equal(t, "10.0.0.0", fakePod.Status.PodIP)
	assert.Equal(t, []v1.PodIP{{IP: "0.0.0.0"}}, fakePod.Status.PodIPs)
}
