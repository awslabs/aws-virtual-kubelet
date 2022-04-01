/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os/exec"

	pb "github.com/aws/aws-virtual-kubelet/proto/vkvmagent/v0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/klog"
)

// GetAgentIdentity returns Instance document from ec2 instance
func (s *server) GetAgentIdentity(ctx context.Context, in *pb.GetAgentIdentityRequest) (*pb.GetAgentIdentityResponse, error) {
	klog.Info("received GetAgentIdentity request", in.String())
	//get instance document from instance
	instanceDoc, err := exec.Command("bash", "-c", "curl http://169.254.169.254/latest/dynamic/instance-identity/document").Output()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	klog.Info("instanceDoc : ", string(instanceDoc))
	doc := pb.EC2InstanceIdentity{InstanceDocument: instanceDoc}
	return &pb.GetAgentIdentityResponse{Ec2InstanceIdentity: &doc}, nil
}

// LaunchAuthenticatedEndpoint creates permanent TLS gRPC connection using the certificates provided in the request.
func (s *server) LaunchAuthenticatedEndpoint(ctx context.Context, in *pb.LaunchAuthenticatedEndpointRequest) (*pb.LaunchAuthenticatedEndpointResponse, error) {
	klog.Info("starting lifeCycleServer ")
	startLifeCycleAgentServer(in.PemCertificateChain, in.PemPrivateKey)
	return &pb.LaunchAuthenticatedEndpointResponse{}, nil
}

// startBootstrapAgentServer runs grpc server with agent bootstrap api's
func startBootstrapAgentServer() error {
	klog.Info("starting bootstrap grpc service")
	userData, err := getUserData()
	if err != nil {
		klog.Error("cannot get userdata: ", err)
	}
	// create Self Signed certification
	certBytes, keyBytes, err := createSelfSignedCert()
	if err != nil {
		klog.Error(err)
		return err
	}

	creds, err := loadBootstrapServerCredentials(certBytes, keyBytes, userData.CACertificate)
	if err != nil {
		klog.Error("cannot load TLS credentials: ", err)
		return err
	}

	lis, err := net.Listen("tcp", bootstrapPort)
	if err != nil {
		klog.Error(err)
		klog.Fatalf("failed to listen: %v", err)
		return err
	}
	s := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterAgentBootstrapServer(s, &server{})
	klog.Info("registered agentBootstrap server, listening on port : ", bootstrapPort)
	if err := s.Serve(lis); err != nil {
		klog.Fatalf("failed to serve BootstrapAgentServer: %v", err)
	}
	return nil
}

// loadBootstrapServerCredentials returns server's Self Signed certification Credentials
func loadBootstrapServerCredentials(certificate []byte, key []byte, CACert string) (credentials.TransportCredentials, error) {
	//load CA cert
	cp := x509.NewCertPool()
	ok := cp.AppendCertsFromPEM([]byte(CACert))
	klog.Info("loaded CA cert ", ok)
	if !ok {
		return nil, fmt.Errorf("failed to add CA's certificate for BootstrapServer ")
	}
	// Load server's certificate and private key
	serverCert, err := tls.X509KeyPair(certificate, key)
	if err != nil {
		return nil, err
	}
	// Create the credentials and return it
	config := &tls.Config{
		ClientCAs:    cp,
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}
	return credentials.NewTLS(config), nil
}
