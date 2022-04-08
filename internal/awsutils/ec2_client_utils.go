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
	"strings"

	b64 "encoding/base64"
	json "encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"k8s.io/klog"
)

// UserData represents the elements of UserData for consumption
type UserData struct {
	VmInit         string `json:"vm-init-config"`
	BootstrapAgent string `json:"bootstrap-agent-config"`
	PresignedURL   string `json:"bootstrap-agent-download-url"`
}

//SecurityGroupNametoID translates a list of SG names (e.g. mySecurityGroup) to SG IDs (e.g. sg-xxxxxxxx)
func (client *Client) SecurityGroupNametoID(ctx context.Context, sgNames []string) (sgIDs []string, err error) {
	input := ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("group-name"),
				Values: sgNames,
			},
		},
	}
	resp, err := client.DescribeSecurityGroups(ctx, &input)
	if err != nil {
		klog.Errorf("unable to describe ec2 Security Groups %v", err)
		return nil, err
	}
	for _, sg := range resp.SecurityGroups {
		sgIDs = append(sgIDs, *sg.GroupId)
	}
	return sgIDs, err
}

// EC2RunInstancesUtil assists in standardizing RunInstancesInput for all VK work
func EC2RunInstancesUtil(
	ctx context.Context,
	IAMInstanceProfile string,
	ImageID string,
	InstanceType string,
	KeyName string,
	SecurityGroupIDs []string,
	SubnetID string,
	Tags []types.TagSpecification,
	UserData string,
	client EC2API) (output *ec2.RunInstancesOutput, err error) {

	var MinCount, MaxCount int32 = 1, 1
	input := ec2.RunInstancesInput{
		IamInstanceProfile: &types.IamInstanceProfileSpecification{
			Name: aws.String(IAMInstanceProfile),
		},
		ImageId:      aws.String(ImageID),
		InstanceType: types.InstanceType(*aws.String(InstanceType)),
		//Monitoring
		MaxCount:          &MaxCount,
		MinCount:          &MinCount,
		SecurityGroupIds:  SecurityGroupIDs,
		SubnetId:          aws.String(SubnetID),
		TagSpecifications: Tags,
		UserData:          aws.String(UserData),
	}
	if KeyName != "" {
		input.KeyName = aws.String(KeyName)
	}
	resp, err := client.RunInstances(ctx, &input)
	return resp, err
}

//GenerateVKVMUserData creates a Base64 encoded string for the intended purpose of use with the EC2 RunInstances API field "UserData"
// Inputs:
//	ctx: Context for all requests in this function.
// 	s3api: The API to attempt to download information from.
// 	bootstrapS3Bucket: S3 Bucket location for Bootstrap Agent e.g. bootstrap-agent-bucket
//  bootstrapS3Key: S3 Key location for Boostrap Agent. e.g. vkvmagent-0.4.0-8
//  VMInit: Instructions to execute on EC2 VM Startup. Downloads bootstrap agent.
//  BootstrapAgent: Instructions to execute after VMInit to startup bootstrap agent.
// Outputs:
//  userdata: base 64 encoded string with presigned URL to download and initialize the bootstrap agent
//  err: any error that might occur as part of attempting to generate UserData
func GenerateVKVMUserData(ctx context.Context, bootstrapS3Bucket string, bootstrapS3Key string, VMInit string, BootstrapAgent string) (userdata string, err error) {
	s3api, err := NewS3Client()
	if err != nil {
		return "", err
	}

	url, err := SignURL(ctx, s3api, &bootstrapS3Bucket, &bootstrapS3Key)
	if err != nil {
		return "", err
	}
	tempData := UserData{
		VmInit:         VMInit,
		BootstrapAgent: BootstrapAgent,
		PresignedURL:   url,
	}
	// stringify UserData struct
	// encode twice, once for Interface reader expecting b64, once for EC2 API Call
	userDataJSON, err := json.Marshal(tempData)
	if err != nil {
		return "", err
	}
	userdata = EncodeUserData(EncodeUserData(replaceHTMLEscapes(string(userDataJSON))))

	return userdata, err
}

func EncodeUserData(userData string) string {
	return b64.StdEncoding.EncodeToString([]byte(userData))
}

// SignURL will generate a presigned URL for a configured location's key and bucket and return it to caller.
func SignURL(ctx context.Context, s3Client S3API, presignBucket *string, presignKey *string) (string, error) {
	input := s3.GetObjectInput{
		Bucket: presignBucket,
		Key:    presignKey,
	}
	resp, err := s3Client.PresignGetObject(ctx, &input)
	if err != nil {
		return "", err
	}
	return resp.URL, nil
}

// replaceHTMLEscapes provides a replacement of HTML escaped characters back to original state.
func replaceHTMLEscapes(stringData string) string {
	stringData = strings.Replace(stringData, "\\u003c", "<", -1)
	stringData = strings.Replace(stringData, "\\u003e", ">", -1)
	stringData = strings.Replace(stringData, "\\u0026", "&", -1)
	return stringData
}
