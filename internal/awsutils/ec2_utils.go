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
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"k8s.io/klog/v2"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-virtual-kubelet/internal/metrics"
	util "github.com/aws/aws-virtual-kubelet/internal/utils"
	corev1 "k8s.io/api/core/v1"
)

// Create Network Interface implementation
// Inputs:
//     tag_value is the value of the Name tag for the Network Interface
//     subnet_id is the id of Subnet where the Network Interface is created
//     eni_description is description of the Network Interface created
// 	   api is the backend service API
// Output:
//     dns is private DNS address of the Network Interface created
//	   ebi_id is the ID of the Network Interface created
//     If success, a nil error.
//     Otherwise, error.

func CreateNetworkInterface(tagValue string, subnetId string, ec2Client EC2API) (privateIp string,
	eniId string, err error) {
	input := &ec2.CreateNetworkInterfaceInput{
		Description: aws.String(tagValue),
		SubnetId:    aws.String(subnetId),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeNetworkInterface,
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(tagValue),
					},
				},
			},
		},
	}
	result, err := ec2Client.CreateNetworkInterface(context.TODO(), input)
	if err != nil {
		metrics.CreateENIErrors.Inc()
		return "", "", errors.New("unable to create Network Interface , tag : " + tagValue + ", " +
			err.Error())
	}
	klog.Info("created a new Network Interface, DNS : ", *result.NetworkInterface.PrivateDnsName)
	return *result.NetworkInterface.PrivateDnsName, *result.NetworkInterface.NetworkInterfaceId, nil
}

// Delete NetworkInterface function implementation
// find the eni id using Name tag and then delete the eni using the id.
// Inputs:
//     tag_value is the value of the Name tag for the Network Interface,
// 	   api is the backend API
// Output:
//     If success, a nil error.
//     Otherwise, error.

func DeleteNetworkInterface(tagValue string, ec2Client EC2API) error {
	// find eni id for the given tag name
	_, eniId, err := GetNetworkInterfaceByTagName(tagValue, ec2Client)
	if err != nil {
		return errors.New("Problem finding the eni id for tag =" + tagValue + " " + err.Error())
	}
	if eniId == "" {
		return nil
	}
	input := &ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: aws.String(eniId),
	}
	_, err = ec2Client.DeleteNetworkInterface(context.TODO(), input)
	if err != nil {
		metrics.DeleteENIErrors.Inc()
		return errors.New("unable to delete Network Interface : " + eniId)
	}
	klog.Info("deleted Network Interface, eniId : ", eniId)
	return nil
}

// Describe NetworkInterface function implementation
// find the eni id and DNS address using Name tag
// Inputs:
//     tag_value is the value of the Name tag for the Network Interface,
// 	   api is the backend API
// Output:
//     dns is the Private DNS address of the Network Interface
//	   eni_id is the id of the Network Interface
//     If success, a nil error.
//     Otherwise, error.

func GetNetworkInterfaceByTagName(tagValue string, ec2Client EC2API) (DNS string, eniId string, err error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: []types.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []string{
					tagValue,
				},
			},
		},
	}
	var dnsOut, eniIdOut string
	result, err := ec2Client.DescribeNetworkInterfaces(context.TODO(), input)
	if err != nil {
		klog.Error(err)
		metrics.DescribeENIErrors.Inc()
		return "", "", errors.New(err.Error() + " , Input tagValue :" + tagValue)
	}
	for _, r := range result.NetworkInterfaces {
		dnsOut = *r.PrivateDnsName
		eniIdOut = *r.NetworkInterfaceId
	}
	klog.Infof("found results - eniId : %v , dns : %v ", eniIdOut, dnsOut)
	return dnsOut, eniIdOut, err
}

// Describe EC2 Instance function implementation
// find the instance status and private ip address per given instance Id
// Inputs:
//     instanceId is the instanceId of the EC2,
// 	   api is the backend API
// Output:
//     status is the status of the EC2 instance
//	   privateIp is the Private IP address of the instance
//     If success, a nil error.
//     Otherwise, error.

func GetInstanceStatusById(instanceId string, ec2Client EC2API) (status string, privateIp string, err error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceId},
	}
	var state, privateIpOut string
	result, err := ec2Client.DescribeInstances(context.TODO(), input)
	if err != nil {
		klog.Error(err)
		metrics.DescribeEC2Errors.Inc()
		return "", "", errors.New(err.Error() + " , Input instanceId :" + instanceId)
	}
	//throw error if number of elements are not equal to 1
	if len(result.Reservations) != 1 {
		klog.Error("DescribeInstances expect 1 Reservations, but got : ", len(result.Reservations), result)
		return "", "", errors.New("DescribeInstances expect 1 Reservations, but got" + strconv.Itoa(len(result.Reservations)))
	}
	if len(result.Reservations[0].Instances) != 1 {
		klog.Error("DescribeInstances expect 1 instance, but got : ", len(result.Reservations[0].Instances), result)
		return "", "", errors.New("DescribeInstances expect 1 instance, but got" + strconv.Itoa(len(result.Reservations[0].Instances)))
	}
	instance := result.Reservations[0].Instances[0]
	if instance.PrivateIpAddress != nil {
		state = string(instance.State.Name)
		privateIpOut = *instance.PrivateIpAddress
	}
	klog.Infof("found results : instanceStatus is : %v  for the given instanceId : %v", state, instanceId)
	return state, privateIpOut, nil
}

// CreateEC2 generates a new EC2 instance based upon the input values provided.
func CreateEC2(ctx context.Context, pod *corev1.Pod, clientTimeoutSeconds int, userData string, presignBucket string, presignKey string) (string, error) {
	ec2Client, err := NewEc2Client(clientTimeoutSeconds)
	if err != nil {
		return "", err
	}

	annotationValue := pod.Annotations["compute.amazonaws.com/tags"]
	annotationTags := make(map[string]string)
	json.Unmarshal([]byte(annotationValue), &annotationTags)
	var tagsInput []types.TagSpecification = []types.TagSpecification{{
		ResourceType: "instance",
		Tags:         []types.Tag{},
	}}
	var tags []types.Tag

	// Loop through tags to unwrangle and assign to []types.TagSpecification
	for key, value := range annotationTags {
		// Append each individual tag to the Tag List
		tags = append(tags, types.Tag{
			Key:   aws.String(strings.Trim(key, " ")),
			Value: aws.String(strings.Trim(value, " ")),
		})
	}
	tagsInput[0].Tags = tags
	if len(tags) == 0 {
		tagsInput = nil
	}

	// Split security group into array of []string
	securityGroups := util.TrimmedStringSplit(pod.Annotations["compute.amazonaws.com/security-groups"], ",")

	// Generate RunInstancesInput
	resp, err := EC2RunInstancesUtil(
		ctx,
		pod.Annotations["compute.amazonaws.com/instance-profile"],
		pod.Annotations["compute.amazonaws.com/image-id"],
		pod.Annotations["compute.amazonaws.com/instance-type"],
		pod.Annotations["compute.amazonaws.com/key-pair"],
		securityGroups,
		pod.Annotations["compute.amazonaws.com/subnet-id"],
		tagsInput,
		userData,
		ec2Client,
	)

	if err != nil {
		klog.Error(err)
		metrics.EC2LaunchErrors.Inc()
		return "", err
	}

	instanceID := *resp.Instances[0].InstanceId

	if instanceID == "" {
		klog.ErrorS(errors.New("InstanceId missing from RunInstances output"), "Pod instance-id Annotation is empty",
			"output", *resp)
	} else {
		klog.InfoS("RunInstances launched an instance", "instance-id", instanceID)
	}

	pod.Annotations["compute.amazonaws.com/instance-id"] = *resp.Instances[0].InstanceId

	metrics.EC2Launched.Inc()
	return *resp.Instances[0].InstanceId, err
}

func TerminateEC2(ctx context.Context, instanceID string, clientTimeoutSeconds int) (string, error) {
	if instanceID == "" {
		return "instance-id-not-set", nil
	}

	terminateInstanceInput := ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}

	ec2Client, err := NewEc2Client(clientTimeoutSeconds)

	resp, err := ec2Client.TerminateInstances(ctx, &terminateInstanceInput)
	if err != nil {
		if strings.Contains(err.Error(), "InvalidInstanceID.NotFound") {
			return "", nil
		}
		metrics.EC2TerminatationErrors.Inc()
		return "", err
	}
	klog.Infof("Terminated EC2 Instance ID : %v", *resp.TerminatingInstances[0].InstanceId)
	metrics.EC2Terminatated.Inc()
	return *resp.TerminatingInstances[0].InstanceId, nil
}

// UpdateInstanceSecurityGroups will modify existing EC2 security groups to list of input SG names
func UpdateInstanceSecurityGroups(ctx context.Context, ec2Client EC2API, instanceID string, sgs []string) (err error) {
	klog.Infof("Assigning new security groups %v to instance %s", sgs, instanceID)
	//Needs to update name to ID for each SG before passing
	if len(sgs) == 0 {
		return nil
	}
	var sgIDs []string
	//Serves to flip to required input style for AWS API if not supplied.
	if !strings.Contains(sgs[0], "sg-") {
		sgIDs, err = ec2Client.SecurityGroupNametoID(ctx, sgs)
		if err != nil {
			klog.Errorf("cannot assign new SGs to instance with error %v", err)
			return err
		}
	} else {
		sgIDs = sgs
	}
	input := ec2.ModifyInstanceAttributeInput{
		InstanceId: &instanceID,
		Groups:     sgIDs,
	}
	_, err = ec2Client.ModifyInstanceAttribute(ctx, &input)
	if err != nil {
		klog.Errorf("unable to update security groups for instance %v to %v", instanceID, sgs)
		return err
	}
	return err
}

// GetPrivateIP gets private ip of the EC2 instance
func GetPrivateIP(instanceID string, clientTimeoutSeconds int) (privateIp string, err error) {
	ec2Client, err := NewEc2Client(clientTimeoutSeconds)
	if err != nil {
		return "", err
	}
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}
	klog.Infof("ec2 instance %v waiting for describe operations", instanceID)

	// begin waiter code.
	err = ec2Client.NewInstanceRunningWaiter(*input)
	if err != nil {
		klog.Errorf("error waiting for instance %v running status , error : %v", instanceID, err)
		return "", errors.New(err.Error() + " , Input instanceId :" + instanceID)
	}
	klog.Infof("ec2 instance %v is now ready for describe operations", instanceID)
	// end waiter code.

	var state, privateIpOut string
	result, err := ec2Client.DescribeInstances(context.TODO(), input)
	if err != nil {
		klog.Error(err)
		return "", errors.New(err.Error() + " , Input instanceId :" + instanceID)
	}
	//throw error if number of reservations / instances are not equal to 1
	if len(result.Reservations) != 1 {
		klog.Error("DescribeInstances call expected 1 Reservations, but got : ", len(result.Reservations), result)
		return "", errors.New("DescribeInstances call expected 1 Reservations, but got : " + strconv.Itoa(len(result.Reservations)))
	}
	if len(result.Reservations[0].Instances) != 1 {
		klog.Error("DescribeInstances call expected 1 instance, but got : ", len(result.Reservations[0].Instances), result)
		return "", errors.New("DescribeInstances expected 1 instance, but got" + strconv.Itoa(len(result.Reservations[0].Instances)))
	}
	instance := result.Reservations[0].Instances[0]
	if instance.PrivateIpAddress != nil {
		state = string(instance.State.Name)
		privateIpOut = *instance.PrivateIpAddress
	} else {
		klog.Errorf("PrivateIpAddress is nil for instance : %v ", instanceID)
		return "", errors.New("PrivateIpAddress is nil for instance : " + instanceID)
	}
	klog.Infof("found results : instanceStatus is : %v , privateIp is : %v, for the given instanceId : %v", state, privateIpOut, instanceID)
	return privateIpOut, nil
}

//UpdateInstanceProfile updates Pod EC2 IAM Association based on input name
//func (p *EC2Provider) UpdateInstanceProfile(ctx context.Context, instanceID string, instanceProfile string) (err error) {
//	klog.Infof("Updating Instance Profile for %s to %s", instanceID, instanceProfile)
//	//https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeIamInstanceProfileAssociations.html
//	// First, get association from above
//	input := ec2.DescribeIamInstanceProfileAssociationsInput{
//		Filters: []types.Filter{{
//			Name:   aws.String("instance-id"),
//			Values: []string{instanceID},
//		},
//			{
//				Name:   aws.String("state"),
//				Values: []string{"associated"},
//			},
//		},
//	}
//	resp, err := p.client.DescribeIamInstanceProfileAssociations(ctx, &input)
//	//https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ReplaceIamInstanceProfileAssociation.html
//	// Then, use above to update association
//	if err != nil {
//		klog.Errorf("unable to describe IAM Instance Profile Associations with error %v", err)
//	}
//	replaceInput := ec2.ReplaceIamInstanceProfileAssociationInput{
//		AssociationId:      resp.IamInstanceProfileAssociations[0].AssociationId,
//		IamInstanceProfile: &types.IamInstanceProfileSpecification{Name: aws.String(instanceProfile)},
//	}
//	_, err = p.client.ReplaceIamInstanceProfileAssociation(ctx, &replaceInput)
//	if err != nil {
//		klog.Error("unable to replace IAM Instance Profile Associations with error ", err)
//	}
//	return err
