// Code generated by smithy-go-codegen DO NOT EDIT.

package ec2

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Enables the specified attachment to propagate routes to the specified
// propagation route table.
func (c *Client) EnableTransitGatewayRouteTablePropagation(ctx context.Context, params *EnableTransitGatewayRouteTablePropagationInput, optFns ...func(*Options)) (*EnableTransitGatewayRouteTablePropagationOutput, error) {
	if params == nil {
		params = &EnableTransitGatewayRouteTablePropagationInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "EnableTransitGatewayRouteTablePropagation", params, optFns, c.addOperationEnableTransitGatewayRouteTablePropagationMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*EnableTransitGatewayRouteTablePropagationOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type EnableTransitGatewayRouteTablePropagationInput struct {

	// The ID of the propagation route table.
	//
	// This member is required.
	TransitGatewayRouteTableId *string

	// Checks whether you have the required permissions for the action, without
	// actually making the request, and provides an error response. If you have the
	// required permissions, the error response is DryRunOperation . Otherwise, it is
	// UnauthorizedOperation .
	DryRun *bool

	// The ID of the attachment.
	TransitGatewayAttachmentId *string

	// The ID of the transit gateway route table announcement.
	TransitGatewayRouteTableAnnouncementId *string

	noSmithyDocumentSerde
}

type EnableTransitGatewayRouteTablePropagationOutput struct {

	// Information about route propagation.
	Propagation *types.TransitGatewayPropagation

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationEnableTransitGatewayRouteTablePropagationMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsEc2query_serializeOpEnableTransitGatewayRouteTablePropagation{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsEc2query_deserializeOpEnableTransitGatewayRouteTablePropagation{}, middleware.After)
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
	if err = addOpEnableTransitGatewayRouteTablePropagationValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opEnableTransitGatewayRouteTablePropagation(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opEnableTransitGatewayRouteTablePropagation(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "ec2",
		OperationName: "EnableTransitGatewayRouteTablePropagation",
	}
}
