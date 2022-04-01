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

	pb "github.com/aws/aws-virtual-kubelet/proto/vkvmagent/v0"
	rpc "github.com/gogo/googleapis/google/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

var podSpec = corev1.PodSpec{}
var successStatus = rpc.Status{Code: 0}

// LaunchApplication implementation
func (s *server) LaunchApplication(ctx context.Context, in *pb.LaunchApplicationRequest) (*pb.LaunchApplicationResponse, error) {
	klog.Infof("received LaunchApplication request for PodSpec: %v", in.GetPod())
	err := createFile(*in.GetPod())
	if err != nil {
		klog.Error(err.Error())
		return nil, err
	}
	return &pb.LaunchApplicationResponse{}, nil
}

// TerminateApplication implementation
func (s *server) TerminateApplication(ctx context.Context, in *pb.TerminateApplicationRequest) (*pb.TerminateApplicationResponse, error) {
	klog.Info("invoked TerminateApplication ")
	return &pb.TerminateApplicationResponse{}, nil
}

// GetApplicationHealth checks application health ( app running on VM )
func (s *server) GetApplicationHealth(ctx context.Context, in *pb.ApplicationHealthRequest) (*pb.ApplicationHealthResponse, error) {
	status := &corev1.PodStatus{Message: "active"}
	return &pb.ApplicationHealthResponse{PodStatus: status}, nil
}

//func (s *server) Check(ctx context.Context, in *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
//	return &pb.HealthCheckResponse{
//		Status: pb.HealthCheckResponse_SERVING,
//	}, nil
//}

// startLifeCycleAgentServer runs grpc server with lifecycle api's
func startLifeCycleAgentServer(serverCert []byte, serverKey []byte) error {
	klog.Info("starting lifecycle agent grpc service")
	userData, err := getUserData()
	if err != nil {
		klog.Error("cannot get userdata: ", err)
	}
	creds, err := loadLifeCycleServerCredentials(serverCert, serverKey, userData.CACertificate)
	if err != nil {
		klog.Error("cannot load TLS credentials: ", err)
		return err
	}

	lis, err := net.Listen("tcp", lifeCyclePort)
	if err != nil {
		klog.Error(err)
		klog.Fatalf("failed to listen: %v", err)
		return err
	}

	s := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterApplicationLifecycleServer(s, &server{})
	klog.Info("registered ApplicaitonLifecycle server listening on ", lifeCyclePort)
	//start server using go routine as we are launching multiple servers
	go func() {
		if err := s.Serve(lis); err != nil {
			klog.Fatalf("failed to serve LifeCycleAgentServer : %v", err)
		}
	}()
	return nil
}

// loadLifeCycleServerCredentials returns server's CA cert Credentials
func loadLifeCycleServerCredentials(serverCert []byte, serverKey []byte, CACert string) (credentials.TransportCredentials, error) {
	//load CA cert
	cp := x509.NewCertPool()
	ok := cp.AppendCertsFromPEM([]byte(CACert))
	klog.Info("loaded CA cert ", ok)
	if !ok {
		return nil, fmt.Errorf("failed to add CA's certificate for ApplicationLifeCycleServer")
	}

	// Load server's certificate and private key
	X509Cert, err := tls.X509KeyPair(serverCert, serverKey)
	if err != nil {
		return nil, err
	}

	// Create the credentials and return it
	config := &tls.Config{
		ClientCAs:    cp,
		Certificates: []tls.Certificate{X509Cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}
	return credentials.NewTLS(config), nil
}
