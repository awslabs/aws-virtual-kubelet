/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package main

import (
	"time"

	pb "github.com/aws/aws-virtual-kubelet/proto/vkvmagent/v0"
	"k8s.io/klog"
)

// server implements UnimplementedApplicationLifecycleServer & UnimplementedAgentBootstrapServer services
type server struct {
	pb.UnimplementedAgentBootstrapServer
	pb.UnimplementedApplicationLifecycleServer
}

// main function launches bootstrapAgent server
func main() {
	klog.Info("main method of grpc service")
	//wait 5 seconds for the VM to boot
	time.Sleep(5 * time.Second)
	startBootstrapAgentServer()
}
