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
	"math/rand"
	"sync"
	"time"

	"github.com/aws/aws-virtual-kubelet/internal/metrics"

	"github.com/aws/aws-virtual-kubelet/internal/config"

	"github.com/aws/aws-virtual-kubelet/internal/awsutils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	corev1 "k8s.io/api/core/v1"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"k8s.io/klog/v2"
)

// Ec2Info contains relevant information of EC2 instances as it pertains to Virtual Kubelet operations.
type Ec2Info struct {
	InstanceID     string   `json:"InstanceID"`
	PrivateIP      string   `json:"PrivateIP"`
	IAMProfile     string   `json:"IAMProfile"`
	SecurityGroups []string `json:"SecurityGroups"`
	RetryCount     int      `json:"RetryCount"`
}

// State is a collection of channels which maintain segments of the Virtual Kubelet's session state.
type State struct {
	ReadyEC2        map[string]Ec2Info `json:"ReadyEC2"`
	ProvisioningEC2 map[string]Ec2Info `json:"ProvisioningEC2"`
	UnhealthyEC2    map[string]Ec2Info `json:"UnhealthyEC2"`
	AllocatedEC2    map[string]Ec2Info `json:"AllocatedEC2"`
	sync.Mutex
}

var (
	// VKState handles the generalized state of a VK process
	VKState State
	//VKHealthState handles the specific state of the healthchecking gRPC calls for VK process
	nodeName                 string
	setPod                   = "set_pod"
	setReady                 = "set_ready"
	setInUse                 = "set_in_use"
	setUnhealthy             = "set_unhealthy" // NOTE not implemented
	initialSetup             = "initial_setup"
	operationPendingWarmpool = "Operation.PENDING_WARMPOOL_PROVISIONING"
	operationReady           = "Operation.Ready"
	operationUnhealthy       = "Operation.Unhealthy"
	operationPendingPod      = "Operation.PENDING_POD_PROVISIONING"
	operationPodInUse        = "Operation.POD_IN_USE"
)

type WarmPoolManager struct {
	config    []config.WarmPoolConfig
	provider  *Ec2Provider
	ec2Client *awsutils.Client
}

func NewWarmPool(ctx context.Context, provider *Ec2Provider) (*WarmPoolManager, error) {
	cfg := config.Config()
	region := cfg.Region

	klog.Infof("Creating Warm Pool Manager with config '%+v'", cfg.WarmPoolConfig)

	ec2Client, err := awsutils.NewEc2Client(region)
	if err != nil {
		klog.Errorf("Can't create Ec2 client in Warm Pool init: %v", err)
		return nil, err
	}

	return &WarmPoolManager{
		config:    cfg.WarmPoolConfig,
		provider:  provider,
		ec2Client: ec2Client,
	}, nil
}

func (wpm *WarmPoolManager) fillAndMaintain() {
	//Generate Initial WarmPool
	if len(wpm.config) > 0 {
		klog.Info("Initializing Warmpool EC2")
		wpm.InitialWarmPoolCreation()

		klog.Info("Starting WarmPool Status Check Ticker")

		go func() {
			ticker := time.NewTicker(60 * time.Second)
			for {
				select {
				case <-ticker.C:
					for _, wpConfig := range wpm.config {
						klog.Infof("Checking Warm Pool depth for config [?]")
						wpm.CheckWarmPoolDepth(context.TODO(), wpConfig)
					}
				}
			}
		}()

		go func() {
			refreshStateTicker := time.NewTicker(60 * time.Minute)
			for {
				select {
				case <-refreshStateTicker.C:
					for range wpm.config {
						klog.Infof("Refreshing Warm Pool status from EC2 tags for config [?]")
						wpm.RefreshWarmPoolFromEC2(context.TODO())
					}
				}
			}
		}()
	}
}

// InitialWarmPoolCreation generates the start-time WarmPool EC2 for Virtual Kubelet
func (wpm *WarmPoolManager) InitialWarmPoolCreation() {
	klog.Info("Generating initial Warmpool Instances")
	wpm.checkEC2TagsForState(context.TODO(), &ec2.DescribeInstancesInput{})
	for _, config := range wpm.config {
		// 	//Check for existing EC2 to import
		existingEC2 := len(VKState.ProvisioningEC2) + len(VKState.ReadyEC2)
		// 	//Submit WarmEC2 function call, loop for DesiredCount
		klog.Infof("Discovered %v existing EC2 for use, creating %v additional", existingEC2, config.DesiredCount-existingEC2)
		for j := existingEC2; j < config.DesiredCount; j++ {
			err := wpm.createWarmEC2(context.TODO(), config)
			if err != nil {
				panic("This createWarmEC2 error wasn't originally handled...determine what to do here")
			}
		}
	}
	return
}

func (wpm *WarmPoolManager) populateEC2Tags(reason string, pod corev1.Pod) (tagSpecification []types.TagSpecification) {
	cfg := config.Config()
	clusterName := cfg.ClusterName

	var tagsInput = []types.TagSpecification{{
		ResourceType: "instance",
		Tags:         []types.Tag{},
	}}
	var tags []types.Tag
	var value string
	if reason == initialSetup {
		value = operationPendingWarmpool
	} else if reason == setReady {
		value = operationReady
	} else if reason == setUnhealthy {
		value = operationUnhealthy
	} else if reason == setPod {
		value = operationPendingPod
		tags = append(tags, types.Tag{
			Key:   aws.String("aws-virtual-kubelet/WarmpoolPodName"),
			Value: aws.String(pod.Name),
		})
		tags = append(tags, types.Tag{
			Key:   aws.String("aws-virtual-kubelet/WarmpoolPodNamespace"),
			Value: aws.String(pod.Namespace),
		})
		tags = append(tags, types.Tag{
			Key:   aws.String("aws-virtual-kubelet/WarmpoolPodUID"),
			Value: aws.String(string(pod.UID)),
		})
		tagsInput[0].Tags = tags
	} else if reason == setInUse {
		value = operationPodInUse
	}
	tagsInput[0].Tags = append(tagsInput[0].Tags,
		types.Tag{
			Key:   aws.String("aws-virtual-kubelet/WarmpoolNodeName"),
			Value: aws.String(nodeName),
		},
		types.Tag{
			Key:   aws.String("aws-virtual-kubelet/WarmpoolClusterName"),
			Value: aws.String(clusterName),
		},
		types.Tag{
			Key:   aws.String("aws-virtual-kubelet/WarmpoolStatus"),
			Value: aws.String(value),
		},
	)
	return tagsInput
}

func (wpm *WarmPoolManager) createWarmEC2(ctx context.Context, wpCfg config.WarmPoolConfig) error {
	klog.Info("Creating Warmpool EC2 Instance")
	tags := wpm.populateEC2Tags(initialSetup, corev1.Pod{})
	instance, privateIP, instanceProfile, securityGroups, err := wpm.CreateWarmEC2(ctx, wpCfg, tags)

	if err != nil {
		klog.Error("Error Creating WarmPool EC2")
		return err
	}
	newSlice := Ec2Info{InstanceID: instance, IAMProfile: instanceProfile, SecurityGroups: securityGroups, RetryCount: 0, PrivateIP: privateIP}
	VKState.ProvisioningEC2[instance] = newSlice
	return err
}

func (wpm *WarmPoolManager) updateEC2Tags(ctx context.Context, instanceID string, reason string, pod corev1.Pod) error {
	tags := wpm.populateEC2Tags(reason, pod)
	input := ec2.CreateTagsInput{}
	input.Resources = append(input.Resources, instanceID)
	input.Tags = tags[0].Tags
	_, err := wpm.ec2Client.CreateTags(ctx, &input)
	if err != nil {
		metrics.EC2TagCreationErrors.Inc()
		klog.ErrorS(err, "error creating ec2 tag", "instanceID", instanceID, "reason", reason, "pod", klog.KObj(&pod))
	}
	return err
}

// CreateWarmEC2 Calls the EC2 RunInstancesAPI with values consistent for a WarmPool using wp and tags as input.
func (wpm *WarmPoolManager) CreateWarmEC2(ctx context.Context, wpConfig config.WarmPoolConfig, tags []types.TagSpecification) (instanceID string, privateIP string, IAMProfileName string, SecurityGroups []string, err error) {
	cfg := config.Config()

	finalUserData, err := awsutils.GenerateVKVMUserData(
		ctx,
		cfg.Region,
		cfg.BootstrapAgent.S3Bucket,
		cfg.BootstrapAgent.S3Key,
		cfg.VMConfig.InitData,
		cfg.BootstrapAgent.InitData,
	)
	if err != nil {
		klog.Errorf("error while creating userdata : %v", err)
	}

	// select a random subnet from the configured list
	s := len(wpConfig.Subnets)
	if s == 0 {
		klog.Error("1 or more Subnets must be configured for Warm Pool...skipping configuration")
		return "", "", "", []string{""}, err
	}
	r := rand.Intn(s) //nolint:gosec
	subnet := wpConfig.Subnets[r]

	klog.Infof("Randomly choosing subnet %v from %v subnets configured for Warm Pool", subnet, s)

	resp, err := awsutils.EC2RunInstancesUtil(
		ctx,
		wpConfig.IamInstanceProfile,
		wpConfig.ImageID,
		wpConfig.InstanceType,
		wpConfig.KeyPair,
		wpConfig.SecurityGroups,
		subnet,
		tags,
		finalUserData,
		wpm.ec2Client,
	)
	if err != nil {
		klog.Errorf("error while generating an EC2 instance: %v", err)
		metrics.WarmEC2LaunchErrors.Inc()
		return "", "", "", []string{""}, err
	}
	klog.Infof("Created EC2 Instance ID in WarmPool: %v", *resp.Instances[0].InstanceId)
	metrics.WarmEC2Launched.Inc()
	instance := resp.Instances[0]

	// collect list of security group ids
	var sgs []string
	for _, sg := range resp.Instances[0].SecurityGroups {
		sgs = append(sgs, *sg.GroupId)
	}

	var instanceProfileID string

	//  handle case where no instance profile is set
	if instance.IamInstanceProfile != nil {
		instanceProfileID = *instance.IamInstanceProfile.Id
	} else {
		instanceProfileID = ""
	}

	return *instance.InstanceId, *instance.PrivateIpAddress, instanceProfileID, sgs, err
}

// CheckWarmPoolDepth Determines the health of the existing WarmPool and then takes appropriate action
func (wpm *WarmPoolManager) CheckWarmPoolDepth(ctx context.Context, wpc config.WarmPoolConfig) {
	klog.Info("Checking WarmPool Depth")
	VKState.Lock()
	defer VKState.Unlock()
	// Check if new EC2 need to be created, or terminated.
	cumulativeWarmEC2 := len(VKState.ReadyEC2) + len(VKState.ProvisioningEC2)
	if (cumulativeWarmEC2) < wpc.DesiredCount {
		for i := 0; i < (wpc.DesiredCount - cumulativeWarmEC2); i++ {
			wpm.createWarmEC2(ctx, wpc)
		}
	} else if (cumulativeWarmEC2) > wpc.DesiredCount {
		var terminatingInstances []string
		var termingInstance Ec2Info
		for i := 0; i < (cumulativeWarmEC2 - wpc.DesiredCount); i++ {
			if len(VKState.ReadyEC2) > 0 {
				termingInstance, VKState.ReadyEC2 = pop(VKState.ReadyEC2)
				terminatingInstances = append(terminatingInstances, termingInstance.InstanceID)
			} else if len(VKState.ProvisioningEC2) > 0 {
				termingInstance, VKState.ProvisioningEC2 = pop(VKState.ProvisioningEC2)
				terminatingInstances = append(terminatingInstances, termingInstance.InstanceID)
			} else {
				klog.Error("Not enough Instances to Terminate, terminating as many as possible")
			}
		}
		klog.Infof("Terminating %v excess WarmPool EC2 Instances", cumulativeWarmEC2-wpc.DesiredCount)
		_, err := wpm.ec2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{InstanceIds: terminatingInstances})
		if err != nil {
			klog.Errorf("unable to terminate warm instances with error : %v", err)
			metrics.WarmEC2TerminationErrors.Inc()
		}
		metrics.WarmEC2Terminated.Inc()
	} else {
		klog.Info("No WarmPool Maintenance Action Taken")
	}
}

// SetNodeName sets NodeName for Tag assignments
func (wpm *WarmPoolManager) SetNodeName(node string) {
	nodeName = node
}

func (wpm *WarmPoolManager) checkEC2TagsForState(ctx context.Context, input *ec2.DescribeInstancesInput) {
	cfg := config.Config()
	clusterName := cfg.ClusterName

	VKState = State{}
	// Initialize the state maps.
	VKState.ReadyEC2 = make(map[string]Ec2Info)
	VKState.ProvisioningEC2 = make(map[string]Ec2Info)
	VKState.UnhealthyEC2 = make(map[string]Ec2Info)
	VKState.AllocatedEC2 = make(map[string]Ec2Info)
	VKState.Lock()
	defer VKState.Unlock()
	first := true
	// Forcibly filter by NodeName to reduce noise
	input.Filters = append(input.Filters, types.Filter{Name: aws.String("tag:aws-virtual-kubelet/WarmpoolNodeName"), Values: []string{nodeName}})
	input.Filters = append(input.Filters, types.Filter{Name: aws.String("tag:aws-virtual-kubelet/WarmpoolClusterName"), Values: []string{clusterName}})
	input.Filters = append(input.Filters, types.Filter{Name: aws.String("instance-state-name"), Values: []string{"running", "pending"}})
	var resp = ec2.DescribeInstancesOutput{}
	for (resp.NextToken != nil) || first {
		first = false
		resp, err := wpm.ec2Client.DescribeInstances(ctx, input)
		if err != nil {
			klog.Errorf("unable to describe instances with error : %v", err)
			return
		}
		if resp.NextToken != nil {
			input.NextToken = resp.NextToken
		}
		for _, reservation := range resp.Reservations {
			for _, instance := range reservation.Instances {
				for i := range instance.Tags {
					if *instance.Tags[i].Key == "aws-virtual-kubelet/WarmpoolStatus" {
						if *instance.Tags[i].Value == operationPendingWarmpool {
							// Move to dedicated function for Updating Tags
							if instance.State.Name == types.InstanceStateNameRunning {
								err := wpm.updateEC2Tags(ctx, *instance.InstanceId, setReady, corev1.Pod{})
								if err != nil {
									klog.Errorf("unable to transition %v from Provisioning to Ready", *instance.InstanceId)
									continue
								}
								VKState.ReadyEC2[*instance.InstanceId] = Ec2Info{InstanceID: *instance.InstanceId, RetryCount: 0, PrivateIP: *instance.PrivateIpAddress}
							} else {
								VKState.ProvisioningEC2[*instance.InstanceId] = Ec2Info{InstanceID: *instance.InstanceId, RetryCount: 0, PrivateIP: *instance.PrivateIpAddress}
							}
						} else if *instance.Tags[i].Value == operationReady {
							VKState.ReadyEC2[*instance.InstanceId] = Ec2Info{InstanceID: *instance.InstanceId, RetryCount: 0, PrivateIP: *instance.PrivateIpAddress}
						} else if *instance.Tags[i].Value == operationUnhealthy {
							VKState.UnhealthyEC2[*instance.InstanceId] = Ec2Info{InstanceID: *instance.InstanceId, RetryCount: 0, PrivateIP: *instance.PrivateIpAddress}
						} else if *instance.Tags[i].Value == operationPendingPod || *instance.Tags[i].Value == operationPodInUse {
							// Added to the Allocated state map. This state map isn't used anywhere. This can be used in the future should an instance assigned to a pod be refurbished.
							// Right now, a pod deletion implies instance termination.
							VKState.AllocatedEC2[*instance.InstanceId] = Ec2Info{InstanceID: *instance.InstanceId, RetryCount: 0, PrivateIP: *instance.PrivateIpAddress}
						}
					}
				}
			}
		}
		klog.Infof("Refresh warmpool completed.")
		klog.Infof("ReadyEC2: %v", VKState.ReadyEC2)
		klog.Infof("ProvisioningEC2: %v", VKState.ProvisioningEC2)
		klog.Infof("UnhealthyEC2: %v", VKState.UnhealthyEC2)
		klog.Infof("AllocatedEC2: %v", VKState.AllocatedEC2)
	}
}

// RefreshWarmPoolFromEC2 reconstructs Virtual Kubelet State from source material of AWS.
func (wpm *WarmPoolManager) RefreshWarmPoolFromEC2(ctx context.Context) {
	klog.Info("Refreshing VK State from EC2 API")
	wpm.checkEC2TagsForState(ctx, &ec2.DescribeInstancesInput{})
}

// GetWarmPoolInstanceIfExist reports if there is an active ready with its instanceID and IP
func (wpm *WarmPoolManager) GetWarmPoolInstanceIfExist(ctx context.Context) (instanceID string, privateIP string, ok bool) {
	klog.Infof("Checking for available Warm Pool instance")

	// Refresh the states first before popping an instance from the Ready state map.
	wpm.RefreshWarmPoolFromEC2(context.TODO())

	VKState.Lock()
	defer VKState.Unlock()
	if len(VKState.ReadyEC2) <= 0 {
		return "", "", false
	}
	var ec2 Ec2Info
	ec2, VKState.ReadyEC2 = pop(VKState.ReadyEC2)
	return ec2.InstanceID, ec2.PrivateIP, true
}

//TerminateInstance provides a way to terminate an EC2 instance.
// To be explicitly used for Warmpool Management and prefer DeletePod once a Pod is set.
func (wpm *WarmPoolManager) TerminateInstance(ctx context.Context, instanceID string) (resp string, err error) {
	cfg := config.Config()
	region := cfg.Region

	resp, err = awsutils.TerminateEC2(ctx, instanceID, region)
	return resp, err
}

//RemoveFromWarmPool finds and removes an EC2 info from WarmPool Cache
func RemoveFromWarmPool(ec2 Ec2Info) (err error) {
	VKState.Lock()
	defer VKState.Unlock()
	delete(VKState.ReadyEC2, ec2.InstanceID)
	delete(VKState.ProvisioningEC2, ec2.InstanceID)
	delete(VKState.UnhealthyEC2, ec2.InstanceID)
	return nil
}

//RemoveFromReadyState mutates Virtual Kubelet state to remove a particular EC2 from VK Ready State.
func RemoveFromReadyState(ec2 Ec2Info) (err error) {
	VKState.Lock()
	defer VKState.Unlock()
	delete(VKState.ReadyEC2, ec2.InstanceID)
	return nil
}

//RemoveFromUnhealthyState mutates Virtual Kubelet state to remove a particular EC2 from VK Ready State.
func RemoveFromUnhealthyState(ec2 Ec2Info) (err error) {
	VKState.Lock()
	defer VKState.Unlock()
	delete(VKState.UnhealthyEC2, ec2.InstanceID)
	return nil
}

// pop returns a random key from the map and mutates the existing map
func pop(stateMap map[string]Ec2Info) (elem Ec2Info, remainder map[string]Ec2Info) {
	key := popKey(stateMap)
	elem = stateMap[key]
	delete(stateMap, key)
	return elem, stateMap
}

// popKey pops a random key in a given map
func popKey(state map[string]Ec2Info) (key string) {
	keys := make([]string, 0, len(state))
	for k := range state {
		keys = append(keys, k)
	}
	return keys[0]
}
