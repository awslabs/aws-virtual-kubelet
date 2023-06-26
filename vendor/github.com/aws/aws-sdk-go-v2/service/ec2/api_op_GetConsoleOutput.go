// Code generated by smithy-go-codegen DO NOT EDIT.

package ec2

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"time"
)

// Gets the console output for the specified instance. For Linux instances, the
// instance console output displays the exact console output that would normally be
// displayed on a physical monitor attached to a computer. For Windows instances,
// the instance console output includes the last three system event log errors. By
// default, the console output returns buffered information that was posted shortly
// after an instance transition state (start, stop, reboot, or terminate). This
// information is available for at least one hour after the most recent post. Only
// the most recent 64 KB of console output is available. You can optionally
// retrieve the latest serial console output at any time during the instance
// lifecycle. This option is supported on instance types that use the Nitro
// hypervisor. For more information, see Instance console output (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-console.html#instance-console-console-output)
// in the Amazon EC2 User Guide.
func (c *Client) GetConsoleOutput(ctx context.Context, params *GetConsoleOutputInput, optFns ...func(*Options)) (*GetConsoleOutputOutput, error) {
	if params == nil {
		params = &GetConsoleOutputInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "GetConsoleOutput", params, optFns, c.addOperationGetConsoleOutputMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*GetConsoleOutputOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type GetConsoleOutputInput struct {

	// The ID of the instance.
	//
	// This member is required.
	InstanceId *string

	// Checks whether you have the required permissions for the action, without
	// actually making the request, and provides an error response. If you have the
	// required permissions, the error response is DryRunOperation . Otherwise, it is
	// UnauthorizedOperation .
	DryRun *bool

	// When enabled, retrieves the latest console output for the instance. Default:
	// disabled ( false )
	Latest *bool

	noSmithyDocumentSerde
}

type GetConsoleOutputOutput struct {

	// The ID of the instance.
	InstanceId *string

	// The console output, base64-encoded. If you are using a command line tool, the
	// tool decodes the output for you.
	Output *string

	// The time at which the output was last updated.
	Timestamp *time.Time

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationGetConsoleOutputMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsEc2query_serializeOpGetConsoleOutput{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsEc2query_deserializeOpGetConsoleOutput{}, middleware.After)
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
	if err = addOpGetConsoleOutputValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opGetConsoleOutput(options.Region), middleware.Before); err != nil {
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

func newServiceMetadataMiddleware_opGetConsoleOutput(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "ec2",
		OperationName: "GetConsoleOutput",
	}
}
