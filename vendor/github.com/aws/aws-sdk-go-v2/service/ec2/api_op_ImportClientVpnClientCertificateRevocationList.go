// Code generated by smithy-go-codegen DO NOT EDIT.

package ec2

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Uploads a client certificate revocation list to the specified Client VPN
// endpoint. Uploading a client certificate revocation list overwrites the existing
// client certificate revocation list. Uploading a client certificate revocation
// list resets existing client connections.
func (c *Client) ImportClientVpnClientCertificateRevocationList(ctx context.Context, params *ImportClientVpnClientCertificateRevocationListInput, optFns ...func(*Options)) (*ImportClientVpnClientCertificateRevocationListOutput, error) {
	if params == nil {
		params = &ImportClientVpnClientCertificateRevocationListInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "ImportClientVpnClientCertificateRevocationList", params, optFns, c.addOperationImportClientVpnClientCertificateRevocationListMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*ImportClientVpnClientCertificateRevocationListOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type ImportClientVpnClientCertificateRevocationListInput struct {

	// The client certificate revocation list file. For more information, see Generate
	// a Client Certificate Revocation List (https://docs.aws.amazon.com/vpn/latest/clientvpn-admin/cvpn-working-certificates.html#cvpn-working-certificates-generate)
	// in the Client VPN Administrator Guide.
	//
	// This member is required.
	CertificateRevocationList *string

	// The ID of the Client VPN endpoint to which the client certificate revocation
	// list applies.
	//
	// This member is required.
	ClientVpnEndpointId *string

	// Checks whether you have the required permissions for the action, without
	// actually making the request, and provides an error response. If you have the
	// required permissions, the error response is DryRunOperation . Otherwise, it is
	// UnauthorizedOperation .
	DryRun *bool

	noSmithyDocumentSerde
}

type ImportClientVpnClientCertificateRevocationListOutput struct {

	// Returns true if the request succeeds; otherwise, it returns an error.
	Return *bool

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationImportClientVpnClientCertificateRevocationListMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsEc2query_serializeOpImportClientVpnClientCertificateRevocationList{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsEc2query_deserializeOpImportClientVpnClientCertificateRevocationList{}, middleware.After)
	if err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddClientRequestIDMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddComputeContentLengthMiddleware(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = v4.AddComputePayloadSHA256Middleware(stack); err != nil {
		return err
	}
	if err = addRetryMiddlewares(stack, options); err != nil {
		return err
	}
	if err = addHTTPSignerV4Middleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addClientUserAgent(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addOpImportClientVpnClientCertificateRevocationListValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opImportClientVpnClientCertificateRevocationList(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecursionDetection(stack); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opImportClientVpnClientCertificateRevocationList(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "ec2",
		OperationName: "ImportClientVpnClientCertificateRevocationList",
	}
}
