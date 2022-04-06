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
	"errors"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	vkconfig "github.com/aws/aws-virtual-kubelet/internal/config"
	"k8s.io/klog"
)

// S3Client Holds a S3 Client and Presign Client
type S3Client struct {
	Svc        *s3.Client
	PresignSvc *s3.PresignClient
}

// NewS3Client Generates a new S3 Client to be utilized across the program.
func NewS3Client() (*S3Client, error) {
	// Initialize client session configuration.
	vkcfg := vkconfig.Config()
	httpClient := http.NewBuildableClient().WithTimeout(time.Second * time.Duration(vkcfg.HealthCheckIntervalSeconds))
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithHTTPClient(httpClient))
	if err != nil {
		klog.Fatalf("unable to load SDK config, %v", err)
		return nil, errors.New("unable to find load SDK config : ")
	}

	svc := s3.NewFromConfig(cfg)
	presignSvc := s3.NewPresignClient(svc)

	// Create the AWS service client.
	return &S3Client{
		Svc:        svc,
		PresignSvc: presignSvc,
	}, nil
}

//PresignGetObject Returns the HTTP response of Getting a presigned URL to a S3 Object
func (client *S3Client) PresignGetObject(ctx context.Context, params *s3.GetObjectInput) (*v4.PresignedHTTPRequest, error) {
	return client.PresignSvc.PresignGetObject(ctx, params)
}
