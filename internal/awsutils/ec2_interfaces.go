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

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// Network Interface API's
type EC2API interface {
	// DescribeNetworkInterfaces retrieves description of Network Interface based on the parameters
	DescribeNetworkInterfaces(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput) (*ec2.DescribeNetworkInterfacesOutput, error)
	// DeleteNetworkInterface deletes Network Interface based on the parameters
	DeleteNetworkInterface(ctx context.Context, params *ec2.DeleteNetworkInterfaceInput) (*ec2.DeleteNetworkInterfaceOutput, error)
	// CreateNetworkInterface creates Network Interface based on the parameters
	CreateNetworkInterface(ctx context.Context, params *ec2.CreateNetworkInterfaceInput) (*ec2.CreateNetworkInterfaceOutput, error)
	// TerminateInstances Terminates EC2 Instances based on params.
	TerminateInstances(ctx context.Context, params *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error)
	// RunInstances Creates EC2 Instances based on input configuration.
	RunInstances(ctx context.Context, input *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error)
	// DescribeInstance retrieves information of EC2 instance based on the parameters
	DescribeInstances(crx context.Context, input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	// CreateTags updates or creates tags of applied resource based on the parameters
	CreateTags(ctx context.Context, input *ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error)
	// ModifyInstanceAttribute modifies existing AWS EC2 Attribute based on input
	ModifyInstanceAttribute(ctx context.Context, input *ec2.ModifyInstanceAttributeInput) (*ec2.ModifyInstanceAttributeOutput, error)
	// SecurityGroupNametoID is a helper that converts SG group names (e.g. "Default") to IDs (e.g. sg-xxxxxx)
	SecurityGroupNametoID(ctx context.Context, sgNames []string) (sgIDs []string, err error)
	// DescribeIamInstanceProfileAssociations describes IAM profile associations for given EC2 instances
	DescribeIamInstanceProfileAssociations(ctx context.Context, input *ec2.DescribeIamInstanceProfileAssociationsInput) (*ec2.DescribeIamInstanceProfileAssociationsOutput, error)
	// ReplaceIamInstanceProfileAssociation describes IAM profile associations for given EC2 instances
	ReplaceIamInstanceProfileAssociation(ctx context.Context, input *ec2.ReplaceIamInstanceProfileAssociationInput) (*ec2.ReplaceIamInstanceProfileAssociationOutput, error)
	//NewInstanceRunningWaiter waits until instance status becomes "running"
	NewInstanceRunningWaiter(input ec2.DescribeInstancesInput) error
}
