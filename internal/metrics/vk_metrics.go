/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc.or its affiliates.All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http: //aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc.or Amazon Web Services EMEA SARL or both.
*/

package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	clientmodel "github.com/prometheus/client_model/go"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	metricsAddr = ":10256"
)

var (
	PodsLaunched = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_pods_created_total",
		Help: "The total number of pods created",
	})
)

var (
	EC2Launched = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_ec2_launched_total",
		Help: "The total number of EC2 instances created using create pod function",
	})
)

var (
	PodsDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_pods_deleted_total",
		Help: "The total number of pods deleted",
	})
)

var (
	GRPCConnectionErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_grpc_connection_errors_total",
		Help: "The total number of GRPC connection errors",
	})
)

var (
	GRPCConnectionTimeouts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_grpc_connection_timeout_errors_total",
		Help: "The total number of GRPC connection timeout errors",
	})
)

var (
	EC2LaunchErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_ec2_launch_errors_total",
		Help: "The total number of errors during EC2 launch",
	})
)

var (
	EC2TerminationErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_ec2_termination_errors_total",
		Help: "The total number of errors during EC2 termination",
	})
)

var (
	EC2Terminated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_ec2_ec2_terminated_total",
		Help: "The total number of EC2 instances terminated using delete pod function",
	})
)

var (
	CreateENIErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_create_eni_errors_total",
		Help: "The total number of errors during network interface creation",
	})
)

var (
	DeleteENIErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_delete_eni_errors_total",
		Help: "The total number of errors during network interface deletion",
	})
)

var (
	DescribeENIErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_describe_eni_errors_total",
		Help: "The total number of errors during describe network interface",
	})
)

var (
	ENICreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_eni_created_total",
		Help: "The total number of network interfaces created",
	})
)
var (
	ENIDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_eni_deleted_total",
		Help: "The total number of network interfaces deleted",
	})
)
var (
	ENIDescribed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_eni_described_total",
		Help: "The total number of times describe network interfaces invoked",
	})
)

var (
	DescribeEC2Errors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_describe_ec2_errors_total",
		Help: "The total number of errors during describe EC2",
	})
)

var (
	LaunchApplicationErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_launch_application_grpc_errors_total",
		Help: "The total number of grpc errors during launch application",
	})
)

var (
	TerminateApplicationErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_terminate_application_grpc_errors_total",
		Help: "The total number of grpc errors during terminate application",
	})
)

var (
	GetAapplicationHealthErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_get_application_health_grpc_errors_total",
		Help: "The total number of grpc errors during get application health",
	})
)

var (
	NodeNameErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_get_nodename_errors_total",
		Help: "The total number of errors during get nodeName operation",
	})
)

var (
	HealthCheckPodCacheError = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_get_pod_from_local_cache_errors_total",
		Help: "The total number of errors during get pod from local vk cache",
	})
)

var (
	HealthCheckGRPCError = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_check_pod_health_grpc_errors_total",
		Help: "The total number of errors during check pod health grpc fun",
	})
)

var (
	MissingHealthCheckResponse = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_check_pod_health_nil_response_total",
		Help: "The total number of nil responses from check pod health grpc call",
	})
)

var (
	HealthCheckPodsDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_pods_deleted_from_health_checks_total",
		Help: "The total number of unhealthy pods deleted from health checks routine",
	})
)

var (
	GetAgentIdentityErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_get_agent_identity_grpc_errors_total",
		Help: "The total number of grpc errors during get agent idenity",
	})
)

var (
	CreateCACertErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_create_ca_cert_errors_total",
		Help: "The total number of errors during CA cert creation",
	})
)

var (
	CreateCertSignedByCACertErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_create_cert_signed_by_ca_errors_total",
		Help: "The total number of errors during creation of server and client certs",
	})
)

var (
	GetSecretErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_retrieve_secret_from_k8s_total",
		Help: "The total number of errors during get secret operation",
	})
)

var (
	CreateSecretErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_create_secret_errors_total",
		Help: "The total number of errors during create secret operation",
	})
)

var (
	UpdateSecretErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_update_secret_errors_total",
		Help: "The total number of errors during update secret operation",
	})
)

var (
	SecretCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_create_secret_from_k8s_total",
		Help: "The total number of a secret is created",
	})
)

var (
	SecretUpdated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_update_secret_from_k8s_total",
		Help: "The total number times a secret is updated",
	})
)

var (
	EmptyCACertCache = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_ca_certs_missing_in_cache_total",
		Help: "The total number times CA cert not found in VK cache",
	})
)

var (
	EmptyClientCertCache = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_client_certs_missing_in_cache_total",
		Help: "The total number times client cert not found in VK cache",
	})
)

var (
	EmptyCACertSecret = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_ca_certs_missing_in_secret_total",
		Help: "The total number times CA cert not found in k8s secret",
	})
)

var (
	ExpiredCACert = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_expired_ca_certs_total",
		Help: "The total number times CA cert is expired",
	})
)

var (
	ExpiredClientCert = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_expired_client_certs_total",
		Help: "The total number times client cert is expired",
	})
)

var (
	HealthCheckStateUnhealthy = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_health_checks_unhealthy_pod_status",
		Help: "The total number of unhealthy pods identified by health checks monitor",
	})
)

var (
	EC2TagCreationErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_ec2_create_tag_errors_total",
		Help: "The total number of errors creating ec2 tags",
	})
)

var (
	WarmEC2TerminationErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_warm_ec2_termination_errors_total",
		Help: "The total number of errors during warm EC2 instances termination",
	})
)

var (
	WarmEC2Launched = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_warm_ec2_launched_total",
		Help: "The total number of warm EC2 instances created",
	})
)

var (
	WarmEC2Terminated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_warm_ec2_terminated_total",
		Help: "The total number of warm EC2 instances terminated",
	})
)

var (
	WarmEC2LaunchErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_warm_ec2_launch_errors_total",
		Help: "The total number of errors during warmpool EC2 launch",
	})
)

var (
	GRPCAppClientErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_grpc_client_errors_total",
		Help: "The total number of errors getting gRPC connection to application lifecycle client",
	})
)

var (
	CheckApplicationHealthErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_check_application_health_errors_total",
		Help: "The total number of grpc errors during check application health",
	})
)

var (
	WatchApplicationHealthErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_watch_application_health_errors_total",
		Help: "The total number of grpc errors during watch application health",
	})
)

var (
	WatchApplicationHealthStreamErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_watch_application_health_stream_errors_total",
		Help: "The total number of grpc errors during watch application health stream",
	})
)

var (
	HealthCheckStateReset = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vkec2_health_checks_reset_pod_status",
		Help: "The total number of times pod unhealthy count is reset to zero",
	})
)

// init() registers all the counters
func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(PodsLaunched)
	metrics.Registry.MustRegister(PodsDeleted)
	metrics.Registry.MustRegister(GRPCConnectionErrors)
	metrics.Registry.MustRegister(GRPCConnectionTimeouts)
	metrics.Registry.MustRegister(EC2LaunchErrors)
	metrics.Registry.MustRegister(EC2TerminationErrors)
	metrics.Registry.MustRegister(CreateENIErrors)
	metrics.Registry.MustRegister(DeleteENIErrors)
	metrics.Registry.MustRegister(DescribeENIErrors)
	metrics.Registry.MustRegister(DescribeEC2Errors)
	metrics.Registry.MustRegister(LaunchApplicationErrors)
	metrics.Registry.MustRegister(TerminateApplicationErrors)
	metrics.Registry.MustRegister(CheckApplicationHealthErrors)
	metrics.Registry.MustRegister(WatchApplicationHealthErrors)
	metrics.Registry.MustRegister(WatchApplicationHealthStreamErrors)
	metrics.Registry.MustRegister(EC2Terminated)
	metrics.Registry.MustRegister(EC2Launched)
	metrics.Registry.MustRegister(NodeNameErrors)
	metrics.Registry.MustRegister(HealthCheckGRPCError)
	metrics.Registry.MustRegister(HealthCheckStateReset)
	metrics.Registry.MustRegister(MissingHealthCheckResponse)
	metrics.Registry.MustRegister(GetAgentIdentityErrors)
	metrics.Registry.MustRegister(CreateCACertErrors)
	metrics.Registry.MustRegister(CreateCertSignedByCACertErrors)
	metrics.Registry.MustRegister(GetSecretErrors)
	metrics.Registry.MustRegister(CreateSecretErrors)
	metrics.Registry.MustRegister(ENICreated)
	metrics.Registry.MustRegister(ENIDeleted)
	metrics.Registry.MustRegister(ENIDescribed)
	metrics.Registry.MustRegister(GRPCAppClientErrors)
	metrics.Registry.MustRegister(WarmEC2LaunchErrors)
	metrics.Registry.MustRegister(WarmEC2Launched)
	metrics.Registry.MustRegister(WarmEC2Terminated)
	metrics.Registry.MustRegister(WarmEC2TerminationErrors)
	metrics.Registry.MustRegister(EC2TagCreationErrors)
	metrics.Registry.MustRegister(HealthCheckStateUnhealthy)
}

// GetMetricsData returns all the metrics for testing purposes
func GetMetricsData() []*clientmodel.MetricFamily {
	gauges, _ := metrics.Registry.Gather()
	klog.Info(" Metrics Gather list (MetricFamily ) = ", gauges)
	return gauges
}

// ExposeMetrics exposes metrics server on VK, use curl http://{vk-ip}:10256/metrics from ec 2 instance to test the endpoint
func ExposeMetrics() {
	// Setup metrics mux.
	vkServerMux := http.NewServeMux()
	vkServerMux.Handle("/metrics", promhttp.Handler())
	vkServerMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})

	metricsServer := &http.Server{
		Addr:         metricsAddr,
		Handler:      vkServerMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	klog.Infof("Listening on %s for metrics and healthz", metricsAddr)
	if err := metricsServer.ListenAndServe(); err != http.ErrServerClosed {
		klog.Fatalf("Error listening: %q", err)
	}
}
