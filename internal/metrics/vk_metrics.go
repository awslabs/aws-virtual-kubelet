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
		Name: "pod_created",
		Help: "The total number of pods created",
	})
)

var (
	EC2Launched = promauto.NewCounter(prometheus.CounterOpts{
		Name: "EC2_launched",
		Help: "The total number of EC2 instances created using create pod function",
	})
)

var (
	PodsDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pods_deleted",
		Help: "The total number of pods deleted",
	})
)

var (
	GRPCConnectionErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "grpc_connection_errors",
		Help: "The total number of GRPC connection errors",
	})
)

var (
	GRPCConnectionTimeouts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "grpc_connection_timeout_errors",
		Help: "The total number of GRPC connection timeout errors",
	})
)

var (
	EC2LaunchErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ec2_launch_errors",
		Help: "The total number of errors during EC2 launch",
	})
)

var (
	EC2TerminatationErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ec2_termination_errors",
		Help: "The total number of errors during EC2 termination",
	})
)

var (
	EC2Terminatated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_ec2_terminated",
		Help: "The total number of EC2 instances terminated using delete pod function",
	})
)

var (
	IPRetrievalErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "IP_retrieval_errors",
		Help: "The total number of errors during retrieval of IP address",
	})
)

var (
	CreateENIErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "create_eni_errors",
		Help: "The total number of errors during network interface creation",
	})
)

var (
	DeleteENIErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "delete_eni_errors",
		Help: "The total number of errors during network interface deletion",
	})
)

var (
	DescribeENIErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "describe_eni_errors",
		Help: "The total number of errors during describe network interface",
	})
)

var (
	DescribeEC2Errors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "describe_ec2_errors",
		Help: "The total number of errors during describe EC2",
	})
)

var (
	LaunchAapplicationErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "launch_application_grpc_errors",
		Help: "The total number of grpc errors during launch application",
	})
)

var (
	TerminateApplicationErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "terminate_application_grpc_errors",
		Help: "The total number of grpc errors during terminate application",
	})
)

var (
	GetAapplicationHealthErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_application_health_grpc_errors",
		Help: "The total number of grpc errors during get application health",
	})
)

var (
	NodeNameErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_nodename_errors",
		Help: "The total number of errors during get nodeName operation",
	})
)

var (
	HealthCheckPodCacheError = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_pod_from_local_cache_error",
		Help: "The total number of errors during get pod from local vk cache",
	})
)

var (
	HealthCheckGRPCError = promauto.NewCounter(prometheus.CounterOpts{
		Name: "check_pod_health_grpc_error",
		Help: "The total number of errors during check pod health grpc fun",
	})
)

var (
	MissingHealthCheckResponse = promauto.NewCounter(prometheus.CounterOpts{
		Name: "check_pod_health_nil_response",
		Help: "The total number of nil responses from check pod health grpc call",
	})
)

var (
	HealthCheckPodsDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pods_deleted_from_health_checks",
		Help: "The total number of unhealthy pods deleted from health checks routine",
	})
)

var (
	GetAgentIdentityErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_agent_identity_grpc_errors",
		Help: "The total number of grpc errors during get agent idenity",
	})
)

var (
	CreateCACertErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "create_ca_cert_errors",
		Help: "The total number of errors during CA cert creation",
	})
)

var (
	CreateCertSignedByCACertErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "create_cert_signed_by_ca_errors",
		Help: "The total number of errors during creation of server and client certs",
	})
)

var (
	GetSecretErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "retrieve_secret_from_k8s",
		Help: "The total number of errors during get secret operation",
	})
)

var (
	CreateSecretErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "create_secret_from_k8s",
		Help: "The total number of errors during create secret operation",
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
	metrics.Registry.MustRegister(EC2TerminatationErrors)
	metrics.Registry.MustRegister(IPRetrievalErrors)
	metrics.Registry.MustRegister(CreateENIErrors)
	metrics.Registry.MustRegister(DeleteENIErrors)
	metrics.Registry.MustRegister(DescribeENIErrors)
	metrics.Registry.MustRegister(DescribeEC2Errors)
	metrics.Registry.MustRegister(LaunchAapplicationErrors)
	metrics.Registry.MustRegister(TerminateApplicationErrors)
	metrics.Registry.MustRegister(GetAapplicationHealthErrors)
	metrics.Registry.MustRegister(EC2Terminatated)
	metrics.Registry.MustRegister(EC2Launched)
	metrics.Registry.MustRegister(NodeNameErrors)
	metrics.Registry.MustRegister(HealthCheckGRPCError)
	metrics.Registry.MustRegister(MissingHealthCheckResponse)
	metrics.Registry.MustRegister(GetAgentIdentityErrors)
	metrics.Registry.MustRegister(CreateCACertErrors)
	metrics.Registry.MustRegister(CreateCertSignedByCACertErrors)
	metrics.Registry.MustRegister(GetSecretErrors)
	metrics.Registry.MustRegister(CreateSecretErrors)
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
