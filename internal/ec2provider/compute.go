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
	"errors"
	"fmt"

	"github.com/aws/aws-virtual-kubelet/internal/config"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/aws/aws-virtual-kubelet/internal/utils"

	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"k8s.io/klog/v2"

	"github.com/aws/aws-virtual-kubelet/internal/awsutils"
	corev1 "k8s.io/api/core/v1"
)

type ComputeManager interface {
	GetCompute(ctx context.Context, pod *corev1.Pod) (string, string, error)
	//createCompute(ctx context.Context, pod *corev1.Pod) (*interface{}, error)
}

type computeManager struct {
	ec2Client *awsutils.Client
}

func NewComputeManager(ctx context.Context) (*computeManager, error) {
	ec2, err := awsutils.NewEc2Client()
	if err != nil {
		return nil, err
	}

	return &computeManager{
		ec2Client: ec2,
	}, nil
}

// GetCompute obtains compute for the given pod.  This compute may come from a Warm Pool, newly created EC2
//  instance, or other appropriate source
func (c *computeManager) GetCompute(ctx context.Context, p *Ec2Provider, pod *corev1.Pod) (string, string, error) {
	// if we already have an instance associated with this pod, just find and return that
	if c.podHasInstance(ctx, pod) {
		podInstanceID := pod.Annotations["compute.amazonaws.com/instance-id"]
		klog.Infof("Pod %v(%v) already assigned to instance %v (reusing compute)",
			pod.Name, pod.Namespace, podInstanceID)
		return podInstanceID, pod.Status.PodIP, nil
		// otherwise, get an instance from warm pool or create one
	}

	// check if pod is configured for Warm Pool (or a default Warm Pool exists)
	if utils.PodIsWarmPool(ctx, pod, len(p.warmPool.config)) {
		klog.Infof("Pod %v(%v) is configured for Warm Pool (or a default Pool is configured)",
			pod.Name, pod.Namespace)
		instanceID, privateIP, instanceFound := p.warmPool.GetWarmPoolInstanceIfExist(ctx)

		if instanceFound {
			pod.Annotations["compute.amazonaws.com/instance-id"] = instanceID
			pod.Status.PodIP = privateIP
			// NOTE pod notification will happen in upstream caller

			// update EC2 tags to mark that the provisioning is in process
			err := p.warmPool.updateEC2Tags(ctx, instanceID, "set_pod", *pod)
			if err != nil {
				klog.ErrorS(err, "Can't update EC2 tags for Warm Pool", "instance",
					instanceID, pod, "pod", klog.KObj(pod))
				return "", "", err
			}

			return instanceID, privateIP, nil
		} else {
			err := errors.New("no instance in 'Ready' state")
			klog.Errorf("Pod %v(%v) is configured to use Warm Pool, but no instance was available: %v",
				pod.Name, pod.Namespace, err)
			return "", "", err
		}
	} else {
		return c.createCompute(ctx, pod)
	}
}

func (c *computeManager) podHasInstance(ctx context.Context, pod *corev1.Pod) bool {
	podInstanceID := pod.Annotations["compute.amazonaws.com/instance-id"]

	// if we already created an instance for this pod (in which case the instance-id annotation will be set)
	if podInstanceID != "" {
		ec2Client, err := awsutils.NewEc2Client()
		if err != nil {
			// if we get an error trying to create an Ec2 client, assume we don't have a valid instance
			return false
		}

		status, err := ec2Client.DescribeInstanceStatus(ctx, &ec2.DescribeInstanceStatusInput{
			// uncomment to include non-Running instances (currently we only look for running instances)
			//IncludeAllInstances: aws.Bool(true),
			InstanceIds: []string{podInstanceID},
		})
		if err != nil || status == nil || len(status.InstanceStatuses) != 1 ||
			status.InstanceStatuses[0].InstanceState == nil {
			// if we get an error trying to describe the instance, or any status info is missing,
			// assume it's invalid and create a new one
			return false
		}
		if status.InstanceStatuses[0].InstanceState.Name == types.InstanceStateNameRunning {
			// assume we received a successful instance status
			klog.Infof("Found (and re-using) existing instance with status %+v", status)
			return true
		}
	}

	// no instance id set in the pod annotation (assume we don't have an instance for this pod then)
	return false
}

// DeleteCompute removes compute for the given pod. NOTE instances are terminated, even if they came from a warm pool
func (c *computeManager) DeleteCompute(ctx context.Context, p *Ec2Provider, pod *corev1.Pod) error {
	return c.deleteCompute(ctx, pod)
}

//func (c *computeManager) createCompute(ctx context.Context, p *Ec2Provider, pod *corev1.Pod) (*interface{}, error) {
func (c *computeManager) createCompute(ctx context.Context, pod *corev1.Pod) (string, string, error) {
	cfg := config.Config()

	klog.Info("Generating a fresh EC2 Instance")
	finalUserData, err := awsutils.GenerateVKVMUserData(
		ctx,
		cfg.BootstrapAgent.S3Bucket,
		cfg.BootstrapAgent.S3Key,
		cfg.VMConfig.InitData,
		cfg.BootstrapAgent.InitData,
	)

	// TODO(guicejg): 🚨 UNDO temporary replacement of user data with startup script
	finalUserData = awsutils.EncodeUserData(cfg.VMConfig.InitData)

	instanceID, err := awsutils.CreateEC2(
		ctx,
		pod,
		finalUserData,
		cfg.BootstrapAgent.S3Bucket,
		cfg.BootstrapAgent.S3Key,
	)

	// Await EC2 Launch
	// NOTE This doesn't wait for EC2 launch, GetPrivateIP below is where the timeout is implemented
	if err != nil {
		return "", "", fmt.Errorf("failed to create ec2 instance, error : %v ", err.Error())
	}

	privateIP, err := awsutils.GetPrivateIP(instanceID)
	if err != nil {
		return "", "", err
	}

	pod.Status.PodIP = privateIP

	return instanceID, privateIP, nil
}

func (c *computeManager) deleteCompute(ctx context.Context, pod *corev1.Pod) error {

	podInstanceID := pod.Annotations["compute.amazonaws.com/instance-id"]

	instanceId, err := awsutils.TerminateEC2(ctx, podInstanceID)
	if err != nil {
		klog.Errorf("error terminating EC2 instance %v: %v", instanceId, err)
	}
	return err
}
