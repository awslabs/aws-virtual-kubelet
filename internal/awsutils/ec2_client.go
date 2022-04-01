/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package awsutils

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"k8s.io/klog"
)

// Client communicates with the regional AWS service.
type Client struct {
	Svc       *ec2.Client
	WaiterSvc *ec2.InstanceRunningWaiter
}

// NewEc2Client creates a new AWS client in the given region.
func NewEc2Client(region string) (*Client, error) {
	// Initialize client session configuration.
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		klog.Fatalf("unable to load SDK config, %v", err)
		return nil, errors.New("unable to find load SDK config : ")
	}
	// Create the AWS service client.
	ec2Client := ec2.NewFromConfig(cfg)
	var options = func(options *ec2.InstanceRunningWaiterOptions) {
		options.MaxDelay = 15 * time.Second
		options.MinDelay = 5 * time.Second
	}

	waiter := ec2.NewInstanceRunningWaiter(ec2Client, options)
	return &Client{
		Svc:       ec2Client,
		WaiterSvc: waiter,
	}, nil
}

// DescribeNetworkInterfaces retrieves description of Network Interface based on the parameters
func (client *Client) DescribeNetworkInterfaces(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput) (*ec2.DescribeNetworkInterfacesOutput, error) {
	return client.Svc.DescribeNetworkInterfaces(ctx, params)
}

// DeleteNetworkInterface deletes Network Interface based on the parameters
func (client *Client) DeleteNetworkInterface(ctx context.Context, params *ec2.DeleteNetworkInterfaceInput) (*ec2.DeleteNetworkInterfaceOutput, error) {
	return client.Svc.DeleteNetworkInterface(ctx, params)
}

// CreateNetworkInterface creates Network Interface based on the parameters
func (client *Client) CreateNetworkInterface(ctx context.Context, params *ec2.CreateNetworkInterfaceInput) (*ec2.CreateNetworkInterfaceOutput, error) {
	return client.Svc.CreateNetworkInterface(ctx, params)
}

// TerminateInstances Terminates EC2 Instances based on params.
func (client *Client) TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
	resp, err := client.Svc.TerminateInstances(ctx, params)
	return resp, err
}

// RunInstances Launches EC2 instances based on RunInstancesInput
func (client *Client) RunInstances(ctx context.Context, input *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error) {
	resp, err := client.Svc.RunInstances(ctx, input)
	return resp, err
}

//DescribeInstance retrieves information of EC2 instance based on the parameters
func (client *Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return client.Svc.DescribeInstances(ctx, params)
}

//DescribeInstanceStatus Describes the status of the specified instances or all of your instances
func (client *Client) DescribeInstanceStatus(ctx context.Context, params *ec2.DescribeInstanceStatusInput) (*ec2.DescribeInstanceStatusOutput, error) {
	return client.Svc.DescribeInstanceStatus(ctx, params)
}

// CreateTags Updates or Assigns tags based on input specifications to AWS resources.
func (client *Client) CreateTags(ctx context.Context, input *ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error) {
	return client.Svc.CreateTags(ctx, input)
}

// ModifyInstanceAttribute modifies existing AWS EC2 Attribute based on input
func (client *Client) ModifyInstanceAttribute(ctx context.Context, input *ec2.ModifyInstanceAttributeInput) (*ec2.ModifyInstanceAttributeOutput, error) {
	return client.Svc.ModifyInstanceAttribute(ctx, input)
}

// DescribeSecurityGroups describes existing AWS EC2 Security Group based on input
func (client *Client) DescribeSecurityGroups(ctx context.Context, input *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	return client.Svc.DescribeSecurityGroups(ctx, input)
}

// DescribeIamInstanceProfileAssociations describes IAM profile associations for given EC2 instances
func (client *Client) DescribeIamInstanceProfileAssociations(ctx context.Context, input *ec2.DescribeIamInstanceProfileAssociationsInput) (*ec2.DescribeIamInstanceProfileAssociationsOutput, error) {
	return client.Svc.DescribeIamInstanceProfileAssociations(ctx, input)
}

// ReplaceIamInstanceProfileAssociation describes IAM profile associations for given EC2 instances
func (client *Client) ReplaceIamInstanceProfileAssociation(ctx context.Context, input *ec2.ReplaceIamInstanceProfileAssociationInput) (*ec2.ReplaceIamInstanceProfileAssociationOutput, error) {
	return client.Svc.ReplaceIamInstanceProfileAssociation(ctx, input)
}

//NewInstanceRunningWaiter waits until instance status becomes "running"
func (client *Client) NewInstanceRunningWaiter(input ec2.DescribeInstancesInput) error {
	waiter := client.WaiterSvc
	maxWaitTime := 60 * time.Second
	err := waiter.Wait(context.TODO(), &input, maxWaitTime)
	if err != nil {
		klog.Errorf("NewInstanceRunningWaiter: %v", err)
		return err
	}
	return nil
}
