package health

import (
	"context"
	"errors"
	"io"
	"log"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	mock_vkvmagent_v0 "github.com/aws/aws-virtual-kubelet/mocks/generated/vkvmagent/v0"
	vkvmagent_v0 "github.com/aws/aws-virtual-kubelet/proto/vkvmagent/v0"

	"github.com/golang/mock/gomock"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-virtual-kubelet/internal/config"

	v1 "k8s.io/api/core/v1"
)

func TestNewCheckResult(t *testing.T) {
	// create minimum config
	_ = config.InitConfig(&config.DirectLoader{DirectConfig: config.ProviderConfig{
		HealthConfig: config.HealthConfig{
			HealthCheckIntervalSeconds: 1,
		},
	}})

	// create fake pod
	fakePod := &v1.Pod{}

	// create monitor for fake pod with check that succeeds after 1s
	m := NewMonitor(fakePod, SubjectVkvma, "UsageTestMonitor", nil)

	type args struct {
		monitor *Monitor
		failed  bool
		message string
		data    interface{}
	}
	tests := []struct {
		name string
		args args
		want *checkResult
	}{
		{
			name: "Failed check result",
			args: args{
				failed:  true,
				message: "FailedCheck",
				data:    nil,
			},
			want: &checkResult{
				Monitor: m,
				Failed:  true,
				Message: "FailedCheck",
				Data:    nil,
			},
		},
		{
			name: "Successful check result with data",
			args: args{
				failed:  true,
				message: "FailedCheck",
				data: &v1.Pod{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "FakePod",
						Namespace: "fake-namespace",
					},
					Spec:   v1.PodSpec{},
					Status: v1.PodStatus{},
				},
			},
			want: &checkResult{
				Monitor: m,
				Failed:  true,
				Message: "FailedCheck",
				Data: &v1.Pod{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "FakePod",
						Namespace: "fake-namespace",
					},
					Spec:   v1.PodSpec{},
					Status: v1.PodStatus{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCheckResult(m, tt.args.failed, tt.args.message, tt.args.data)

			// ensure timestamps always match for this test
			tt.want.Timestamp = got.Timestamp

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCheckResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonitor_usage(t *testing.T) {
	// create minimum config
	_ = config.InitConfig(&config.DirectLoader{DirectConfig: config.ProviderConfig{
		HealthConfig: config.HealthConfig{
			HealthCheckIntervalSeconds: 1,
		},
	}})

	// create fake pod
	fakePod := &v1.Pod{}

	// create monitor for fake pod with check that succeeds after 1s
	m := NewMonitor(fakePod, SubjectVkvma, "UsageTestMonitor", nil)

	assert.Equal(t, MonitoringStateUnknown, string(m.State))

	m.check = func(ctx context.Context, m *Monitor) *checkResult {
		// wait 1s and return success
		time.Sleep(time.Second * 1)

		return NewCheckResult(m, false, "SuccessfulCheck", nil)
	}

	// create a fake channel to simulate check handler receive
	m.handlerReceiver = make(chan *checkResult)

	// create cancellation context to stop monitor goroutines
	ctx, cancel := context.WithCancel(context.Background())

	// create wait group to wait for monitor goroutine to stop
	waitGroup := sync.WaitGroup{}

	// start monitor (with wait group)
	m.Run(ctx, &waitGroup)

	// simulate some other activity (monitor will run 3 checks during this time)
	time.Sleep(time.Second * 3)

	// cancel the monitor
	cancel()

	// wait for and assert that the monitor goroutine ended
	waitGroup.Wait()
	assert.False(t, m.IsMonitoring, "Monitoring should be stopped.")
}

func TestMonitorHandlerReception(t *testing.T) {
	// NOTE same as TestMonitor_usage but start a handler receiver goroutine to exercise that portion of monitor code

	// create minimum config
	_ = config.InitConfig(&config.DirectLoader{DirectConfig: config.ProviderConfig{
		HealthConfig: config.HealthConfig{
			HealthCheckIntervalSeconds: 1,
		},
	}})

	// create fake pod
	fakePod := &v1.Pod{}

	// create monitor for fake pod with check that succeeds after 1s
	m := NewMonitor(fakePod, SubjectVkvma, "TestMonitor", nil)

	m.check = func(ctx context.Context, m *Monitor) *checkResult {
		// wait 1s and return success
		time.Sleep(time.Second * 1)

		return NewCheckResult(m, false, "SuccessfulCheck", nil)
	}

	// create a fake channel to simulate check handler receive
	m.handlerReceiver = make(chan *checkResult)

	// create cancellation context to stop monitor goroutines
	ctx, cancel := context.WithCancel(context.Background())

	// create wait group to wait for monitor goroutine to stop
	waitGroup := sync.WaitGroup{}

	// start monitor (with wait group)
	m.Run(ctx, &waitGroup)

	// start receiver simulator (NOTE could use a non-blocking channel but a goroutine approximates the actual code)
	go func(in chan *checkResult) {
		for result := range in {
			// receive and print results
			log.Printf("Check handler simulation received result: %+v", result)
		}
	}(m.handlerReceiver)

	// ensure there is enough time for at least one result to be received
	time.Sleep(time.Second * 2)

	// cancel the monitor
	cancel()

	// wait for and assert that the monitor goroutine ended
	waitGroup.Wait()
	assert.False(t, m.IsMonitoring, "Monitoring should be stopped.")
}

func TestMonitor_incrementFailures(t *testing.T) {
	type fields struct {
		Resource     interface{}
		Subject      Subject
		Name         string
		Failures     int
		State        MonitoringState
		IsMonitoring bool
		check        func(ctx context.Context, monitor *Monitor) *checkResult
		isWatcher    bool
		stream       func(ctx context.Context, monitor *Monitor) interface{}
	}
	type args struct {
		unhealthyThreshold int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   MonitoringState
	}{
		{
			name: "Monitor becomes unhealthy when failure threshold is _met_",
			fields: fields{
				Failures: 1,
				State:    MonitoringStateUnknown,
			},
			args: args{
				unhealthyThreshold: 1,
			},
			want: MonitoringStateUnhealthy,
		},
		{
			name: "Monitor becomes unhealthy when failure threshold is _exceeded_",
			fields: fields{
				Failures: 2,
				State:    MonitoringStateUnknown,
			},
			args: args{
				unhealthyThreshold: 1,
			},
			want: MonitoringStateUnhealthy,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create fake pod
			fakePod := &v1.Pod{}

			// create monitor for fake pod with check that succeeds after 1s
			m := NewMonitor(fakePod, SubjectVkvma, "TestMonitor", nil)

			m.incrementFailures(tt.args.unhealthyThreshold)

			if got := m.State; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("m.State = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonitor_resetFailures(t *testing.T) {
	type fields struct {
		Resource        interface{}
		Subject         Subject
		Name            string
		Failures        int
		State           MonitoringState
		IsMonitoring    bool
		check           func(ctx context.Context, monitor *Monitor) *checkResult
		isWatcher       bool
		stream          func(ctx context.Context, monitor *Monitor) interface{}
		handlerReceiver chan *checkResult
	}
	tests := []struct {
		name   string
		fields fields
		want   MonitoringState
	}{
		{
			name: "Monitor becomes healthy when failures are reset",
			fields: fields{
				Failures: 5,
				State:    MonitoringStateUnhealthy,
			},
			want: MonitoringStateHealthy,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create fake pod
			fakePod := &v1.Pod{}

			// create monitor for fake pod with check that succeeds after 1s
			m := NewMonitor(fakePod, SubjectVkvma, "TestMonitor", nil)

			m.resetFailures()

			if got := m.State; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("m.State = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonitorWatch(t *testing.T) {
	// create minimum config
	_ = config.InitConfig(&config.DirectLoader{DirectConfig: config.ProviderConfig{
		//HealthConfig: config.HealthConfig{
		//	HealthCheckIntervalSeconds: 1,
		//},
	}})

	// create fake pod
	fakePod := &v1.Pod{}

	// create monitor for fake pod with check that succeeds after 1s
	m := NewMonitor(fakePod, SubjectVkvma, "TestMonitor", nil)

	assert.Equal(t, MonitoringStateUnknown, string(m.State))

	m.isWatcher = true

	// create a fake channel to simulate check handler receive
	m.handlerReceiver = make(chan *checkResult)
	// start receiver simulator (NOTE could use a non-blocking channel but a goroutine approximates the actual code)
	go func(in chan *checkResult) {
		for result := range in {
			// receive and print results
			log.Printf("Check handler simulation received result: %+v", result)
		}
	}(m.handlerReceiver)

	// create cancellation context to stop monitor goroutines
	ctx, cancel := context.WithCancel(context.Background())

	// setup mocks
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create wait group to wait for monitor goroutine to stop
	waitGroup := sync.WaitGroup{}

	// set stream function (that returns a streaming gRPC client with a .Recv() function)
	m.getStream = func(ctx context.Context, monitor *Monitor) interface{} {
		// create mock clients
		appClient := mock_vkvmagent_v0.NewMockApplicationLifecycleClient(ctrl)
		streamClient := mock_vkvmagent_v0.NewMockApplicationLifecycle_WatchApplicationHealthClient(ctrl)

		// expect a call to watch app health and return a stream client
		appClient.EXPECT().
			WatchApplicationHealth(ctx, &vkvmagent_v0.ApplicationHealthRequest{}).
			Return(streamClient, nil)

		// NOTE the returned client is missing .EXPECT() so we use the streamClient below to expect another call
		stream, err := appClient.WatchApplicationHealth(ctx, &vkvmagent_v0.ApplicationHealthRequest{})
		if err != nil {
			panic("This mock should not return an error")
		}

		// expect a call to Recv(), sleep to simulate activity, then return a PodStatus inside an App Health Response
		streamClient.EXPECT().Recv().DoAndReturn(func() (*vkvmagent_v0.ApplicationHealthResponse, error) {
			time.Sleep(3 * time.Second)
			podStatus := &vkvmagent_v0.ApplicationHealthResponse{PodStatus: &v1.PodStatus{}}
			return podStatus, nil
		})

		return stream
	}

	// start monitor (with wait group)
	m.Run(ctx, &waitGroup)

	// simulate some other activity (monitor will run 3 checks during this time)
	time.Sleep(time.Second * 1)

	// cancel the monitor
	cancel()

	// wait for and assert that the monitor goroutine ended
	waitGroup.Wait()
	assert.False(t, m.IsMonitoring, "Monitoring should be stopped.")
}

func TestMonitorWatchEOFRecovery(t *testing.T) {
	// create minimum config
	_ = config.InitConfig(&config.DirectLoader{DirectConfig: config.ProviderConfig{
		HealthConfig: config.HealthConfig{
			StreamRetryIntervalSeconds: 1,
		},
	}})

	// create fake pod
	fakePod := &v1.Pod{}

	// create monitor for fake pod with check that succeeds after 1s
	m := NewMonitor(fakePod, SubjectVkvma, "TestMonitor", nil)

	assert.Equal(t, MonitoringStateUnknown, string(m.State))

	m.isWatcher = true

	var checksHandled uint32

	// create a fake channel to simulate check handler receive
	m.handlerReceiver = make(chan *checkResult)
	// start receiver simulator (NOTE could use a non-blocking channel but a goroutine approximates the actual code)
	go func(in chan *checkResult) {
		for result := range in {
			// receive and print results
			log.Printf("Check handler simulation received result: %+v", result)
			atomic.AddUint32(&checksHandled, 1)
		}
	}(m.handlerReceiver)

	// create cancellation context to stop monitor goroutines
	ctx, cancel := context.WithCancel(context.Background())

	// setup mocks
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create wait group to wait for monitor goroutine to stop
	waitGroup := sync.WaitGroup{}

	// track number of times getStream is called to allow different behavior over time
	getStreamCalls := 0

	// set stream function (that returns a streaming gRPC client with a .Recv() function)
	m.getStream = func(ctx context.Context, monitor *Monitor) interface{} {
		getStreamCalls++

		// create mock clients
		appClient := mock_vkvmagent_v0.NewMockApplicationLifecycleClient(ctrl)
		streamClient := mock_vkvmagent_v0.NewMockApplicationLifecycle_WatchApplicationHealthClient(ctrl)

		// expect a call to watch app health and return a stream client
		appClient.EXPECT().
			WatchApplicationHealth(ctx, &vkvmagent_v0.ApplicationHealthRequest{}).
			Return(streamClient, nil)

		// NOTE the returned client is missing .EXPECT() so we use the streamClient below to expect another call
		stream, err := appClient.WatchApplicationHealth(ctx, &vkvmagent_v0.ApplicationHealthRequest{})
		if err != nil {
			panic("This mock should not return an error")
		}

		if getStreamCalls == 1 {
			// expect a call to Recv(), sleep to simulate activity, then return an EOF
			streamClient.EXPECT().Recv().DoAndReturn(func() (*vkvmagent_v0.ApplicationHealthResponse, error) {
				time.Sleep(1 * time.Second)
				return nil, io.EOF
			})
		} else {
			// expect a call to Recv(), return valid responses every second
			streamClient.EXPECT().Recv().AnyTimes().Do(func() { time.Sleep(1 * time.Second) }).
				Return(&vkvmagent_v0.ApplicationHealthResponse{PodStatus: &v1.PodStatus{}}, nil)
		}

		return stream
	}

	// start monitor (with wait group)
	m.Run(ctx, &waitGroup)

	// simulate some other activity to allow some checks to process
	time.Sleep(time.Second * 3)

	assert.True(t, atomic.LoadUint32(&checksHandled) > 0, "checks were successfully handled following an EOF")

	// cancel the monitor
	cancel()

	// wait for and assert that the monitor goroutine ended
	waitGroup.Wait()
	assert.False(t, m.IsMonitoring, "Monitoring should be stopped.")
}

func TestMonitorWatchErrorRecovery(t *testing.T) {
	// create minimum config
	_ = config.InitConfig(&config.DirectLoader{DirectConfig: config.ProviderConfig{
		HealthConfig: config.HealthConfig{
			StreamRetryIntervalSeconds: 1,
		},
	}})

	// create fake pod
	fakePod := &v1.Pod{}

	// create monitor for fake pod with check that succeeds after 1s
	m := NewMonitor(fakePod, SubjectVkvma, "TestMonitor", nil)

	assert.Equal(t, MonitoringStateUnknown, string(m.State))

	m.isWatcher = true

	var checksHandled uint32

	// create a fake channel to simulate check handler receive
	m.handlerReceiver = make(chan *checkResult)
	// start receiver simulator (NOTE could use a non-blocking channel but a goroutine approximates the actual code)
	go func(in chan *checkResult) {
		for result := range in {
			// receive and print results
			log.Printf("Check handler simulation received result: %+v", result)
			atomic.AddUint32(&checksHandled, 1)
		}
	}(m.handlerReceiver)

	// create cancellation context to stop monitor goroutines
	ctx, cancel := context.WithCancel(context.Background())

	// setup mocks
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create wait group to wait for monitor goroutine to stop
	waitGroup := sync.WaitGroup{}

	// track number of times getStream is called to allow different behavior over time
	getStreamCalls := 0

	// set stream function (that returns a streaming gRPC client with a .Recv() function)
	m.getStream = func(ctx context.Context, monitor *Monitor) interface{} {
		getStreamCalls++

		// create mock clients
		appClient := mock_vkvmagent_v0.NewMockApplicationLifecycleClient(ctrl)
		streamClient := mock_vkvmagent_v0.NewMockApplicationLifecycle_WatchApplicationHealthClient(ctrl)

		// expect a call to watch app health and return a stream client
		appClient.EXPECT().
			WatchApplicationHealth(ctx, &vkvmagent_v0.ApplicationHealthRequest{}).
			Return(streamClient, nil)

		// NOTE the returned client is missing .EXPECT() so we use the streamClient below to expect another call
		stream, err := appClient.WatchApplicationHealth(ctx, &vkvmagent_v0.ApplicationHealthRequest{})
		if err != nil {
			panic("This mock should not return an error")
		}

		if getStreamCalls == 1 {
			// expect a call to Recv(), sleep to simulate activity, then return an EOF
			streamClient.EXPECT().Recv().DoAndReturn(func() (*vkvmagent_v0.ApplicationHealthResponse, error) {
				time.Sleep(1 * time.Second)
				return nil, errors.New("👹")
			})
		} else { // expect a call to Recv(), return valid responses every second
			streamClient.EXPECT().Recv().AnyTimes().Do(func() { time.Sleep(1 * time.Second) }).
				Return(&vkvmagent_v0.ApplicationHealthResponse{PodStatus: &v1.PodStatus{}}, nil)
		}

		return stream
	}

	// start monitor (with wait group)
	m.Run(ctx, &waitGroup)

	// simulate some other activity to allow some checks to process
	time.Sleep(time.Second * 3)

	assert.True(t, atomic.LoadUint32(&checksHandled) > 0, "checks were successfully handled following an EOF")

	// cancel the monitor
	cancel()

	// wait for and assert that the monitor goroutine ended
	waitGroup.Wait()
	assert.False(t, m.IsMonitoring, "Monitoring should be stopped.")
}

func TestStopMonitorWithNilStream(t *testing.T) {
	// create minimum config
	_ = config.InitConfig(&config.DirectLoader{DirectConfig: config.ProviderConfig{
		//HealthConfig: config.HealthConfig{
		//	HealthCheckIntervalSeconds: 1,
		//},
	}})

	// create fake pod
	fakePod := &v1.Pod{}

	// create monitor for fake pod with check that succeeds after 1s
	m := NewMonitor(fakePod, SubjectVkvma, "TestMonitor", nil)

	assert.Equal(t, MonitoringStateUnknown, string(m.State))

	m.isWatcher = true

	// create a fake channel to simulate check handler receive
	m.handlerReceiver = make(chan *checkResult)
	// start receiver simulator (NOTE could use a non-blocking channel but a goroutine approximates the actual code)
	go func(in chan *checkResult) {
		for result := range in {
			// receive and print results
			log.Printf("Check handler simulation received result: %+v", result)
		}
	}(m.handlerReceiver)

	// create cancellation context to stop monitor goroutines
	ctx, cancel := context.WithCancel(context.Background())

	// setup mocks
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create wait group to wait for monitor goroutine to stop
	waitGroup := sync.WaitGroup{}

	// set stream function (that returns a streaming gRPC client with a .Recv() function)
	m.getStream = func(ctx context.Context, monitor *Monitor) interface{} {
		return nil
	}

	// start monitor (with wait group)
	m.Run(ctx, &waitGroup)

	// simulate some other activity (monitor will run 3 checks during this time)
	time.Sleep(time.Second * 2)

	// cancel the monitor
	cancel()

	// wait for and assert that the monitor goroutine ended
	waitGroup.Wait()
	assert.False(t, m.IsMonitoring, "Monitoring should be stopped.")
}
