/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseks"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/jsii-runtime-go"

	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	// "github.com/aws/jsii-runtime-go"
)

type VKStackProps struct {
	awscdk.StackProps
}

func NewVKStack(scope constructs.Construct, id string, props *VKStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// EKS Cluster
	cluster := awseks.NewCluster(stack, jsii.String("VKCluster"), &awseks.ClusterProps{
		Version:     awseks.KubernetesVersion_V1_19(),
		ClusterName: jsii.String("virtual-kubelet-cluster"), // Overrides the auto-generated name
	})

	// Virtual Kubelet Namespace manifest
	namespaceManifest := map[string]interface{}{
		"apiVersion":"v1",
		"kind":"Namespace",
		"metadata": map[string]interface{}{
			"name": "virtual-kubelet",
			"labels": map[string]interface{}{
				"name": "virtual-kubelet",
			},
		},
	}

	// Add Namespace to cluster (CDK does not currently have a higher-level construct for doing this in Go)
	namespace := cluster.AddManifest(jsii.String("VKNamespace"),&namespaceManifest)

	// EKS Service Account (IAM Role and associated Kubernetes Service Account)
	// See https://github.com/aws/aws-cdk/tree/master/packages/%40aws-cdk/aws-eks#service-accounts
	serviceAccount := cluster.AddServiceAccount(jsii.String("VKServiceAccount"), &awseks.ServiceAccountOptions{
		Name:      jsii.String("virtual-kubelet-sa"),
		Namespace: jsii.String("virtual-kubelet"),
	})

	// Ensure the Namespace is created before the Service Account that goes in it
	serviceAccount.Node().AddDependency(namespace)

	// Allow EKS (Virtual Kubelet) Service Account to manage EC2 instances
	serviceAccount.Role().AddToPrincipalPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions: jsii.Strings(
			// permission use / justification noted in comments below
			"ec2:CreateNetworkInterface", // needed to create the ENI that serves as the virtual kubelet "node"
			"ec2:DescribeNetworkInterfaces", // needed to obtain the private IP of the ENI for node-naming
			"ec2:DeleteNetworkInterface", // needed to remove the "node" ENI
			"ec2:RunInstances", // needed to launch pod and warm pool instances
			"ec2:DescribeInstances", // needed to get EC2 information and status
			"ec2:TerminateInstances",  // needed to remove pod and warm pool instances
			"ec2:CreateTags", // needed to tag pod and warm pool instances
		),
		// TODO add Tag or other conditions to limit `DeleteNetworkInterface` and `TerminateInstances` to those created
		//  by virtual-kubelet
		//Conditions:    nil,
		Effect:    awsiam.Effect_ALLOW,
		Resources: jsii.Strings("*"),
	}))

	// ECR repository for VK docker images
	awsecr.NewRepository(stack, jsii.String("VKEcr"), &awsecr.RepositoryProps{
		RepositoryName: jsii.String("aws-virtual-kubelet"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("VPC"), &awscdk.CfnOutputProps{
		Value:       cluster.Vpc().VpcId(),
		Description: jsii.String("VPC ID for the EKS cluster"),
	})

	// create an S3 bucket for pipeline use (e.g. staging the VKVMAgent binary)
	s3Bucket := awss3.NewBucket(stack, jsii.String("S3Bucket"), &awss3.BucketProps{
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		// NOTE you will probably need to choose another name for this bucket since they are globally unique.
		//  ⚠️  Be sure to also update any references elsewhere in this project (e.g. sample user data script)
		BucketName:    jsii.String("vk-general-assets"),
		Encryption:    awss3.BucketEncryption_S3_MANAGED,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY, // only destroyed if empty
		Versioned:     jsii.Bool(false),
	})

	awsec2.NewSecurityGroup(stack, jsii.String("VKSecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc:               cluster.Vpc(),
		AllowAllOutbound:  jsii.Bool(true),
		Description:       jsii.String("Default security group for virtual-kubelet pods"),
		SecurityGroupName: jsii.String("virtual-kubelet-default"),
	})

	s3ReadOnlyPolicy := awsiam.ManagedPolicy_FromManagedPolicyArn(stack,
		jsii.String("VKS3ReadOnlyPolicy"), jsii.String("arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"))
	managedPolicies := []awsiam.IManagedPolicy{s3ReadOnlyPolicy}

	instanceRole := awsiam.NewRole(stack, jsii.String("VKInstanceRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ec2.amazonaws.com"),
			&awsiam.ServicePrincipalOpts{}),
		Description:     jsii.String("Default virtual-kubelet instance role"),
		ManagedPolicies: &managedPolicies,
		RoleName:        jsii.String("virtual-kubelet-instance-role"),
	})
	awsiam.NewCfnInstanceProfile(stack, jsii.String("VKInstanceProfile"), &awsiam.CfnInstanceProfileProps{
		Roles:               jsii.Strings(*instanceRole.RoleName()),
		InstanceProfileName: jsii.String("virtual-kubelet-instance-profile"),
	})

	// enable the VKP service account role to pass the instance role to created EC2 instances
	instanceRole.GrantPassRole(serviceAccount)

	// Allow the launched pod EC2 instance role to read the VKVMAgent from S3
	s3Bucket.GrantRead(instanceRole, "*") // may want to specify an S3 key also

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewVKStack(app, "VKStack", &VKStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
