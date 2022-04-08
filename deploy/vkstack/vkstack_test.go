/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package main

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	assertions "github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
)

func TestPipelineInfraStack(t *testing.T) {
	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewVKStack(app, "TestStack", nil)

	// THEN
	template := assertions.Template_FromStack(stack)

	// RESOURCE COUNT VALIDATIONS

	template.ResourceCountIs(jsii.String("AWS::EC2::VPC"), jsii.Number(1))
	template.ResourceCountIs(jsii.String("Custom::AWSCDK-EKS-Cluster"), jsii.Number(1))
	template.ResourceCountIs(jsii.String("Custom::AWSCDK-EKS-KubernetesResource"), jsii.Number(3))

	// RESOURCE PROPERTY VALIDATIONS

	// ensure public buckets are disabled
	template.HasResourceProperties(jsii.String("AWS::S3::Bucket"), map[string]interface{}{
		"PublicAccessBlockConfiguration": map[string]interface{}{
			"BlockPublicAcls":       jsii.Bool(true),
			"BlockPublicPolicy":     jsii.Bool(true),
			"IgnorePublicAcls":      jsii.Bool(true),
			"RestrictPublicBuckets": jsii.Bool(true),
		},
	})
}
