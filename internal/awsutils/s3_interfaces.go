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

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3API defines the interface for S3 API actions for either Mock or official AWS Client usage.
type S3API interface {
	// GetObject Retrieves objects from Amazon S3.
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput) (*v4.PresignedHTTPRequest, error)
}
