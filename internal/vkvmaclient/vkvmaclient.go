/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package vkvmaclient

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-virtual-kubelet/internal/config"

	vkvmagent "github.com/aws/aws-virtual-kubelet/proto/vkvmagent/v0"

	"google.golang.org/grpc/connectivity"

	health "github.com/aws/aws-virtual-kubelet/proto/grpc/health/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"k8s.io/klog"
)

type GrpcClient interface {
	Connect(ctx context.Context) (*grpc.ClientConn, error)
	GetHealthClient(ctx context.Context) (health.HealthClient, error)
	GetApplicationLifecycleClient(ctx context.Context) (vkvmagent.ApplicationLifecycleClient, error)
	IsConnected(ctx context.Context) bool
}

type VkvmaClient struct {
	VkvmaConnection
	health.HealthClient
	vkvmagent.ApplicationLifecycleClient
}

type VkvmaConnection struct {
	config        config.VkvmaConfig
	address       string
	port          int
	connection    *grpc.ClientConn
	serviceConfig string
}

func NewVkvmaClient(ip string, port int) *VkvmaClient {
	cfg := config.Config().VKVMAgentConnectionConfig
	klog.Infof("VKVMA Client loaded cfg %+v", cfg)

	vkvmaConnection := &VkvmaConnection{
		address:       ip,
		port:          port,
		serviceConfig: "",
		config:        cfg,
	}

	return &VkvmaClient{
		VkvmaConnection: *vkvmaConnection,
	}
}

func (v *VkvmaClient) GetHealthClient(ctx context.Context) (health.HealthClient, error) {
	// attempt to connect VKVMA if not already connected
	if !v.IsConnected(ctx) {
		_, err := v.Connect(ctx)
		if err != nil {
			return nil, err
		}
	}
	v.HealthClient = health.NewHealthClient(v.VkvmaConnection.connection)
	return v.HealthClient, nil
}

func (v *VkvmaClient) GetApplicationLifecycleClient(ctx context.Context) (vkvmagent.ApplicationLifecycleClient, error) {
	// attempt to connect VKVMA if not already connected
	if !v.IsConnected(ctx) {
		_, err := v.Connect(ctx)
		if err != nil {
			return nil, err
		}
	}
	v.ApplicationLifecycleClient = vkvmagent.NewApplicationLifecycleClient(v.VkvmaConnection.connection)
	return v.ApplicationLifecycleClient, nil
}

func (v *VkvmaClient) IsConnected(ctx context.Context) bool {

	if v.VkvmaConnection.connection == nil {
		return false
	}
	return v.VkvmaConnection.connection.GetState() == connectivity.Ready
}

func (v *VkvmaClient) Connect(ctx context.Context) (*grpc.ClientConn, error) {
	timeout := time.Duration(v.config.TimeoutSeconds) * time.Second

	ctx, cancel := context.WithTimeout(ctx, timeout)

	defer cancel()

	dialAddr := fmt.Sprintf("%v:%v", v.address, v.port)

	klog.Infof("initiating gRPC connection to %v", dialAddr)

	connStart := time.Now()

	connectParams := grpc.ConnectParams{
		Backoff: backoff.Config{
			BaseDelay:  time.Duration(v.config.Backoff.BaseDelaySeconds) * time.Second,
			Multiplier: v.config.Backoff.Multiplier,
			Jitter:     v.config.Backoff.Jitter,
			MaxDelay:   time.Duration(v.config.Backoff.MaxDelaySeconds) * time.Second,
		},
		MinConnectTimeout: time.Duration(v.config.MinConnectTimeoutSeconds) * time.Second,
	}

	clientParams := keepalive.ClientParameters{
		Time:    time.Duration(v.config.Keepalive.TimeSeconds) * time.Second,
		Timeout: time.Duration(v.config.Keepalive.TimeoutSeconds) * time.Second,
		// If true, client sends keepalive pings even with no active RPCs. If false,
		// when there are no active RPCs, Time and Timeout will be ignored and no
		// keepalive pings will be sent.
		//PermitWithoutStream: true,
	}

	conn, err := grpc.DialContext(
		ctx,
		dialAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithConnectParams(connectParams),
		grpc.WithKeepaliveParams(clientParams),
	)

	select {
	case <-time.After(1 * time.Second):
		klog.Fatalf("connecting to %v pending...(%v timeout)", dialAddr, timeout)
	case <-ctx.Done():
		err := ctx.Err()
		connDuration := time.Since(connStart)
		klog.Errorf("unable to connect to %v after %v:%v", dialAddr, connDuration, err)
	default:
		connDuration := time.Since(connStart)
		klog.Infof("connection established after %v", connDuration)
		v.connection = conn
	}

	if err != nil {
		connDuration := time.Since(connStart)
		klog.Errorf("unable to connect to %v after %v", dialAddr, connDuration)
		return nil, err
	}

	klog.Infof("VKVMAgent server connection setup complete")

	return conn, nil
}
