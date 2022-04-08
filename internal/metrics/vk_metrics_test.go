/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package metrics

import (
	"strconv"
	"strings"
	"testing"

	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

// TestMetricsCounter tests the pod created metric
func TestMetricsCounter(t *testing.T) {
	cases := []struct {
		podCounter int
	}{
		{
			podCounter: 1,
		},
	}
	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			//increment pod created metric
			PodsLaunched.Inc()
			//get all metrics data
			metricFamilyList := GetMetricsData()
			//find PodCreated metrics value from the MetricsFamily list
			var podMetric io_prometheus_client.MetricFamily
			for i := 0; i < len(metricFamilyList); i++ {
				if strings.Compare(*metricFamilyList[i].Name, "vkec2_pods_created_total") == 0 {
					podMetric = *metricFamilyList[i]
					break
				}
			}
			assert.Equal(t, int(*podMetric.Metric[0].Counter.Value), tt.podCounter)
		})
	}
}
