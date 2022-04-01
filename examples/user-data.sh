#!/bin/bash

# This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
# © 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.
#
# This AWS Content is provided subject to the terms of the AWS Customer Agreement
# available at http://aws.amazon.com/agreement or other written agreement between
# Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.

# The following script is an example using EC2 User Data to fetch and run the VKVMAgent binary on EC2 first run
# This example assumes an EC2 Instance Role has been configured along with an installation of the AWS CLI to enable
# access to the otherwise private S3 bucket.  The provided CDK pipeline example sets up an appropriate instance role
# and using an Amazon Linux EC2 image will result in the AWS CLI being available at boot time without additional steps.

# To prepare your script for use as userdata, run the equivalent to the following command in your environment:
# `cat userData.sh | base64`
# Capture the resulting string and update the InitData key in the example config map.

# NOTE Creating a service that will persist through reboots is OS-dependent and outside the scope of this example

aws s3 cp s3://my-vk-bucket/vkvmagent .
chmod u+x vkvmagent
./vkvmagent
