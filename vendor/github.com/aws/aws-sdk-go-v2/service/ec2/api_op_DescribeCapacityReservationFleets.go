// Code generated by smithy-go-codegen DO NOT EDIT.

package ec2

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Describes one or more Capacity Reservation Fleets.
func (c *Client) DescribeCapacityReservationFleets(ctx context.Context, params *DescribeCapacityReservationFleetsInput, optFns ...func(*Options)) (*DescribeCapacityReservationFleetsOutput, error) {
	if params == nil {
		params = &DescribeCapacityReservationFleetsInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "DescribeCapacityReservationFleets", params, optFns, c.addOperationDescribeCapacityReservationFleetsMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*DescribeCapacityReservationFleetsOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type DescribeCapacityReservationFleetsInput struct {

	// The IDs of the Capacity Reservation Fleets to describe.
	CapacityReservationFleetIds []string

	// Checks whether you have the required permissions for the action, without
	// actually making the request, and provides an error response. If you have the
	// required permissions, the error response is DryRunOperation . Otherwise, it is
	// UnauthorizedOperation .
	DryRun *bool

	// One or more filters.
	//   - state - The state of the Fleet ( submitted | modifying | active |
	//   partially_fulfilled | expiring | expired | cancelling | cancelled | failed ).
	//   - instance-match-criteria - The instance matching criteria for the Fleet. Only
	//   open is supported.
	//   - tenancy - The tenancy of the Fleet ( default | dedicated ).
	//   - allocation-strategy - The allocation strategy used by the Fleet. Only
	//   prioritized is supported.
	Filters []types.Filter

	// The maximum number of results to return for the request in a single page. The
	// remaining results can be seen by sending another request with the returned
	// nextToken value. This value can be between 5 and 500. If maxResults is given a
	// larger value than 500, you receive an error.
	MaxResults *int32

	// The token to use to retrieve the next page of results.
	NextToken *string

	noSmithyDocumentSerde
}

type DescribeCapacityReservationFleetsOutput struct {

	// Information about the Capacity Reservation Fleets.
	CapacityReservationFleets []types.CapacityReservationFleet

	// The token to use to retrieve the next page of results. This value is null when
	// there are no more results to return.
	NextToken *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationDescribeCapacityReservationFleetsMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsEc2query_serializeOpDescribeCapacityReservationFleets{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsEc2query_deserializeOpDescribeCapacityReservationFleets{}, middleware.After)
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
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opDescribeCapacityReservationFleets(options.Region), middleware.Before); err != nil {
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

// DescribeCapacityReservationFleetsAPIClient is a client that implements the
// DescribeCapacityReservationFleets operation.
type DescribeCapacityReservationFleetsAPIClient interface {
	DescribeCapacityReservationFleets(context.Context, *DescribeCapacityReservationFleetsInput, ...func(*Options)) (*DescribeCapacityReservationFleetsOutput, error)
}

var _ DescribeCapacityReservationFleetsAPIClient = (*Client)(nil)

// DescribeCapacityReservationFleetsPaginatorOptions is the paginator options for
// DescribeCapacityReservationFleets
type DescribeCapacityReservationFleetsPaginatorOptions struct {
	// The maximum number of results to return for the request in a single page. The
	// remaining results can be seen by sending another request with the returned
	// nextToken value. This value can be between 5 and 500. If maxResults is given a
	// larger value than 500, you receive an error.
	Limit int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// DescribeCapacityReservationFleetsPaginator is a paginator for
// DescribeCapacityReservationFleets
type DescribeCapacityReservationFleetsPaginator struct {
	options   DescribeCapacityReservationFleetsPaginatorOptions
	client    DescribeCapacityReservationFleetsAPIClient
	params    *DescribeCapacityReservationFleetsInput
	nextToken *string
	firstPage bool
}

// NewDescribeCapacityReservationFleetsPaginator returns a new
// DescribeCapacityReservationFleetsPaginator
func NewDescribeCapacityReservationFleetsPaginator(client DescribeCapacityReservationFleetsAPIClient, params *DescribeCapacityReservationFleetsInput, optFns ...func(*DescribeCapacityReservationFleetsPaginatorOptions)) *DescribeCapacityReservationFleetsPaginator {
	if params == nil {
		params = &DescribeCapacityReservationFleetsInput{}
	}

	options := DescribeCapacityReservationFleetsPaginatorOptions{}
	if params.MaxResults != nil {
		options.Limit = *params.MaxResults
	}

	for _, fn := range optFns {
		fn(&options)
	}

	return &DescribeCapacityReservationFleetsPaginator{
		options:   options,
		client:    client,
		params:    params,
		firstPage: true,
		nextToken: params.NextToken,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *DescribeCapacityReservationFleetsPaginator) HasMorePages() bool {
	return p.firstPage || (p.nextToken != nil && len(*p.nextToken) != 0)
}

// NextPage retrieves the next DescribeCapacityReservationFleets page.
func (p *DescribeCapacityReservationFleetsPaginator) NextPage(ctx context.Context, optFns ...func(*Options)) (*DescribeCapacityReservationFleetsOutput, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params
	params.NextToken = p.nextToken

	var limit *int32
	if p.options.Limit > 0 {
		limit = &p.options.Limit
	}
	params.MaxResults = limit

	result, err := p.client.DescribeCapacityReservationFleets(ctx, &params, optFns...)
	if err != nil {
		return nil, err
	}
	p.firstPage = false

	prevToken := p.nextToken
	p.nextToken = result.NextToken

	if p.options.StopOnDuplicateToken &&
		prevToken != nil &&
		p.nextToken != nil &&
		*prevToken == *p.nextToken {
		p.nextToken = nil
	}

	return result, nil
}

func newServiceMetadataMiddleware_opDescribeCapacityReservationFleets(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "ec2",
		OperationName: "DescribeCapacityReservationFleets",
	}
}
