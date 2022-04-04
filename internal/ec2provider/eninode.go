/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package ec2provider

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-virtual-kubelet/internal/config"

	"github.com/aws/aws-virtual-kubelet/internal/metrics"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/resource"

	"k8s.io/klog/v2"

	"github.com/aws/aws-virtual-kubelet/internal/awsutils"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
)

type EniNode struct {
	name               string
	hostname           string
	lastTransitionTime time.Time
}

func GetOrCreateEniNode(ctx context.Context, clientTimeoutSeconds int, clusterName string, subnetId string) (*EniNode, error) {
	ec2Client, err := awsutils.NewEc2Client(clientTimeoutSeconds)
	if err != nil {
		return nil, err
	}

	var eniTag string

	// append unique, consistent value to ENI tag to allow multiple VKP processes to coexist
	vkPodName := os.Getenv("POD_NAME")

	if vkPodName != "" {
		eniTag = clusterName + "-" + vkPodName
	} else {
		klog.Warningf("⚠️  To run more than one instance of this Virtual Kubelet provider, you must set a " +
			"POD_NAME environment variable that is unique (but consistent) per instance.  When deploying using the example " +
			"stateful set, Kubernetes will set this value automatically")
		eniTag = clusterName
	}

	nodeName, err := getOrCreateNodeName(eniTag, subnetId, ec2Client)
	if err != nil {
		return nil, err
	}

	return &EniNode{
		name:               nodeName,
		hostname:           nodeName,
		lastTransitionTime: time.Now(),
	}, nil
}

// getOrCreateNodeName gets private dns name of the network interface for the given parameters (creating one if needed)
func getOrCreateNodeName(tagValue string, subnetId string, ec2Client awsutils.EC2API) (string, error) {
	if tagValue == "" || subnetId == "" {
		return "", errors.New("Parameters tagValue, subnetId  are required. ")
	}

	klog.Infof("Fetching ENI by Name Tag value '%v'", tagValue)
	var privateIP, _, err = awsutils.GetNetworkInterfaceByTagName(tagValue, ec2Client)
	if err != nil {
		metrics.NodeNameErrors.Inc()
		return "", err
	}
	if privateIP == "" {
		klog.Infof("ENI with Name Tag value '%v' not found, creating...", tagValue)
		privateIP, _, _ = awsutils.CreateNetworkInterface(tagValue, subnetId, ec2Client)
		klog.Infof("ENI created, private IP address: '%v'", privateIP)
	}

	// NOTE This fargate prefix must exist for interoperability with EKS
	vkPodName := "fargate-" + privateIP

	return vkPodName, nil
}

func (en *EniNode) Configure(ctx context.Context, k8sNode *corev1.Node) (*corev1.Node, error) {
	k8sNode.Name = en.name
	k8sNode.Labels = map[string]string{
		"type":                   "virtual-kubelet",
		"kubernetes.io/role":     "agent",
		"beta.kubernetes.io/os":  strings.ToLower(config.DefaultOperatingSystem),
		"kubernetes.io/hostname": en.hostname,
		"alpha.service-controller.kubernetes.io/exclude-balancer": "true",
	}
	systemInfo := corev1.NodeSystemInfo{
		KubeletVersion:  "",
		OperatingSystem: config.DefaultOperatingSystem,
		Architecture:    "amd64",
	}

	defaultCapacity := corev1.ResourceList{
		"cpu":     resource.MustParse(config.DefaultCpuCapacity),
		"memory":  resource.MustParse(config.DefaultMemoryCapacity),
		"storage": resource.MustParse(config.DefaultStorageCapacity),
		"pods":    resource.MustParse(config.DefaultPodCapacity),
	}

	lastHeartbeatTime := v1.Now()
	lastTransitionTime := v1.NewTime(en.lastTransitionTime)
	lastTransitionReason := "Virtual Kubelet is ready"
	lastTransitionMessage := "ok"

	readyConditions := []corev1.NodeCondition{
		{
			Type:               corev1.NodeReady,
			Status:             corev1.ConditionTrue,
			LastHeartbeatTime:  lastHeartbeatTime,
			LastTransitionTime: lastTransitionTime,
			Reason:             lastTransitionReason,
			Message:            lastTransitionMessage,
		},
		{
			Type:               corev1.NodePIDPressure,
			Status:             corev1.ConditionFalse,
			LastHeartbeatTime:  lastHeartbeatTime,
			LastTransitionTime: lastTransitionTime,
			Reason:             lastTransitionReason,
			Message:            lastTransitionMessage,
		},
		{
			Type:               corev1.NodeMemoryPressure,
			Status:             corev1.ConditionFalse,
			LastHeartbeatTime:  lastHeartbeatTime,
			LastTransitionTime: lastTransitionTime,
			Reason:             lastTransitionReason,
			Message:            lastTransitionMessage,
		},
		{
			Type:               corev1.NodeDiskPressure,
			Status:             corev1.ConditionFalse,
			LastHeartbeatTime:  lastHeartbeatTime,
			LastTransitionTime: lastTransitionTime,
			Reason:             lastTransitionReason,
			Message:            lastTransitionMessage,
		},
		{
			Type:               corev1.NodeNetworkUnavailable,
			Status:             corev1.ConditionFalse,
			LastHeartbeatTime:  lastHeartbeatTime,
			LastTransitionTime: lastTransitionTime,
			Reason:             lastTransitionReason,
			Message:            lastTransitionMessage,
		},
		{
			Type:               "KubeletConfigOk",
			Status:             corev1.ConditionTrue,
			LastHeartbeatTime:  lastHeartbeatTime,
			LastTransitionTime: lastTransitionTime,
			Reason:             lastTransitionReason,
			Message:            lastTransitionMessage,
		},
	}

	k8sNode.Status = corev1.NodeStatus{
		Capacity:    defaultCapacity,
		Allocatable: defaultCapacity,
		//Phase:           "",
		Conditions: readyConditions,
		Addresses:  nil,
		//DaemonEndpoints: corev1.NodeDaemonEndpoints{},
		NodeInfo: systemInfo,
		//Images:          nil,
		//VolumesInUse:    nil,
		//VolumesAttached: nil,
		//Config:          nil,
	}

	return k8sNode, nil
}

// NodeProvider methods (required)
// See https://pkg.go.dev/github.com/virtual-kubelet/virtual-kubelet/node#NodeProvider

func (en *EniNode) Ping(ctx context.Context) error {
	panic("What should this do? Check that ENI exists and/or is reachable?")
}

func (en *EniNode) NotifyNodeStatus(ctx context.Context, cb func(*corev1.Node)) {
	panic("Save the passed-in callback and use it to notify VK of any node status changes")
}
