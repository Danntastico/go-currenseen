package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
)

// DynamoDBRepository implements the ExchangeRateRepository interface using AWS DynamoDB.
// This is an adapter in the Hexagonal Architecture pattern, connecting the domain layer
// to the AWS DynamoDB infrastructure.
type DynamoDBRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBRepository creates a new DynamoDB repository.
//
// This constructor follows Go best practices and enables dependency injection.
// The client can be a real DynamoDB client or a mock for testing.
//
// Parameters:
//   - client: The DynamoDB client (can be real or mock)
//   - tableName: The name of the DynamoDB table to use
//
// Returns a new DynamoDBRepository instance.
func NewDynamoDBRepository(client *dynamodb.Client, tableName string) *DynamoDBRepository {
	return &DynamoDBRepository{
		client:    client,
		tableName: tableName,
	}
}

// dynamoItem represents a DynamoDB item structure.
// This struct is used for marshaling/unmarshaling between Go and DynamoDB AttributeValue format.
// The dynamodbav tags tell the AWS SDK how to map struct fields to DynamoDB attributes.
type dynamoItem struct {
	PK        string  `dynamodbav:"PK"`            // Partition key: RATE#USD#EUR
	Base      string  `dynamodbav:"Base"`          // Base currency code (e.g., "USD")
	Target    string  `dynamodbav:"Target"`        // Target currency code (e.g., "EUR")
	Rate      float64 `dynamodbav:"Rate"`          // Exchange rate value
	Timestamp int64   `dynamodbav:"Timestamp"`     // Unix timestamp in seconds
	Stale     bool    `dynamodbav:"Stale"`         // Whether rate is marked as stale
	TTL       *int64  `dynamodbav:"ttl,omitempty"` // TTL timestamp (Unix epoch in seconds), optional
}

// entityToDynamoItem converts a domain entity to DynamoDB item format.
//
// This function:
// - Builds the partition key from base and target currencies
// - Converts time.Time to Unix timestamp (seconds)
// - Calculates TTL timestamp from time.Duration
// - Converts CurrencyCode to string
//
// The ttl parameter is used to calculate when the item should expire.
// If ttl is 0 or negative, no TTL is set (TTL will be nil).
func entityToDynamoItem(rate *entity.ExchangeRate, ttl time.Duration) (*dynamoItem, error) {
	if rate == nil {
		return nil, fmt.Errorf("exchange rate cannot be nil")
	}

	// Calculate TTL timestamp (Unix epoch in seconds)
	// DynamoDB TTL requires Unix timestamp in seconds
	var ttlTimestamp *int64
	if ttl > 0 {
		ttlSec := time.Now().Add(ttl).Unix()
		ttlTimestamp = &ttlSec
	}

	return &dynamoItem{
		PK:        buildPartitionKey(rate.Base, rate.Target),
		Base:      rate.Base.String(),
		Target:    rate.Target.String(),
		Rate:      rate.Rate,
		Timestamp: rate.Timestamp.Unix(),
		Stale:     rate.Stale,
		TTL:       ttlTimestamp,
	}, nil
}

// dynamoItemToEntity converts a DynamoDB item to domain entity.
//
// This function:
// - Validates currency codes using domain validation
// - Converts Unix timestamp back to time.Time
// - Creates a new ExchangeRate entity with validation
//
// Returns an error if the stored data is invalid (e.g., invalid currency codes).
// This provides data integrity - even if corrupted data is stored, we validate it.
func dynamoItemToEntity(item *dynamoItem) (*entity.ExchangeRate, error) {
	if item == nil {
		return nil, fmt.Errorf("dynamo item cannot be nil")
	}

	// Validate and create currency codes (domain validation)
	base, err := entity.NewCurrencyCode(item.Base)
	if err != nil {
		return nil, fmt.Errorf("invalid base currency in stored data: %w", err)
	}

	target, err := entity.NewCurrencyCode(item.Target)
	if err != nil {
		return nil, fmt.Errorf("invalid target currency in stored data: %w", err)
	}

	// Convert Unix timestamp back to time.Time
	timestamp := time.Unix(item.Timestamp, 0)

	// Create domain entity with validation
	return entity.NewExchangeRate(base, target, item.Rate, timestamp, item.Stale)
}

// buildPartitionKey creates a partition key from currency codes.
//
// Format: RATE#{BASE}#{TARGET}
// Example: RATE#USD#EUR, RATE#GBP#JPY
//
// This format:
// - Makes the key type explicit (RATE# prefix)
// - Enables direct lookup for Get() and Delete() operations
// - Follows DynamoDB best practices for composite keys
func buildPartitionKey(base, target entity.CurrencyCode) string {
	return fmt.Sprintf("RATE#%s#%s", base.String(), target.String())
}

// marshalDynamoItem converts a dynamoItem to DynamoDB AttributeValue map.
// This is used when writing items to DynamoDB (PutItem).
func marshalDynamoItem(item *dynamoItem) (map[string]types.AttributeValue, error) {
	if item == nil {
		return nil, fmt.Errorf("dynamo item cannot be nil")
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dynamo item: %w", err)
	}

	return av, nil
}

// unmarshalDynamoItem converts a DynamoDB AttributeValue map to dynamoItem.
// This is used when reading items from DynamoDB (GetItem, Query).
func unmarshalDynamoItem(av map[string]types.AttributeValue) (*dynamoItem, error) {
	if av == nil {
		return nil, fmt.Errorf("attribute value map cannot be nil")
	}

	var item dynamoItem
	err := attributevalue.UnmarshalMap(av, &item)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal dynamo item: %w", err)
	}

	return &item, nil
}

// mapDynamoDBError maps DynamoDB errors to domain errors or wraps them appropriately.
//
// This function:
// - Preserves context cancellation errors (returns them as-is)
// - Maps DynamoDB-specific errors to domain errors where appropriate
// - Wraps other errors with context for debugging
//
// Note: Item not found is handled separately by checking result.Item == nil,
// not through error mapping, as DynamoDB GetItem returns success with nil item.
func mapDynamoDBError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// Check for context cancellation - return as-is
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	// Check for DynamoDB resource not found (table/index doesn't exist)
	var resourceNotFoundErr *types.ResourceNotFoundException
	if errors.As(err, &resourceNotFoundErr) {
		// This typically means table or index doesn't exist
		// Return wrapped error with context
		return fmt.Errorf("dynamodb resource not found (%s): %w", operation, err)
	}

	// For other errors, wrap with operation context
	return fmt.Errorf("dynamodb %s failed: %w", operation, err)
}

// Get retrieves an exchange rate for a specific currency pair.
//
// This method:
// - Builds the partition key from base and target currencies
// - Uses GetItem for direct lookup by partition key
// - Returns entity.ErrRateNotFound if the rate doesn't exist
// - Returns rates regardless of TTL expiration (use cases handle expiration)
//
// Context cancellation: Returns error if ctx is cancelled.
func (r *DynamoDBRepository) Get(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
	// Check context before starting operation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Build partition key
	pk := buildPartitionKey(base, target)

	// Prepare GetItem input
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
		},
	}

	// Execute GetItem
	result, err := r.client.GetItem(ctx, input)
	if err != nil {
		return nil, mapDynamoDBError(err, "get item")
	}

	// Check if item exists
	if result.Item == nil {
		return nil, entity.ErrRateNotFound
	}

	// Unmarshal DynamoDB item to dynamoItem
	item, err := unmarshalDynamoItem(result.Item)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal dynamodb item: %w", err)
	}

	// Convert dynamoItem to domain entity
	return dynamoItemToEntity(item)
}

// Save stores an exchange rate with TTL.
//
// This method:
// - Uses PutItem operation which creates or updates (upsert behavior)
// - Converts domain entity to DynamoDB item format
// - Calculates and stores TTL timestamp
// - Marshals item to DynamoDB AttributeValue format
//
// If the rate already exists, it will be updated with new values.
// The TTL is calculated from the current time plus the provided ttl duration.
//
// Context cancellation: Returns error if ctx is cancelled.
func (r *DynamoDBRepository) Save(ctx context.Context, rate *entity.ExchangeRate, ttl time.Duration) error {
	// Check context before starting operation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Convert entity to DynamoDB item (includes TTL calculation)
	item, err := entityToDynamoItem(rate, ttl)
	if err != nil {
		return fmt.Errorf("failed to convert entity to dynamo item: %w", err)
	}

	// Marshal to DynamoDB AttributeValue map
	av, err := marshalDynamoItem(item)
	if err != nil {
		return fmt.Errorf("failed to marshal dynamo item: %w", err)
	}

	// Prepare PutItem input (upsert behavior)
	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	}

	// Execute PutItem
	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		return mapDynamoDBError(err, "put item")
	}

	return nil
}

// GetByBase retrieves all exchange rates for a base currency.
//
// This method:
// - Uses Query operation (not Scan) for efficiency
// - Queries the GSI BaseCurrencyIndex by base currency
// - Returns all rates for the specified base currency
// - Returns empty slice (not nil) if no rates are found
// - Returns rates regardless of TTL expiration (use cases handle expiration)
//
// Context cancellation: Returns error if ctx is cancelled.
func (r *DynamoDBRepository) GetByBase(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
	// Check context before starting operation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Prepare Query input for GSI
	// Note: "Base" is a reserved keyword in DynamoDB, so we use ExpressionAttributeNames
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("BaseCurrencyIndex"),
		KeyConditionExpression: aws.String("#base = :base"),
		ExpressionAttributeNames: map[string]string{
			"#base": "Base", // Map #base to the actual attribute name "Base"
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":base": &types.AttributeValueMemberS{Value: base.String()},
		},
	}

	// Execute Query
	result, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, mapDynamoDBError(err, "query")
	}

	// Convert items to entities
	// Pre-allocate slice with capacity for better performance
	rates := make([]*entity.ExchangeRate, 0, len(result.Items))
	for _, item := range result.Items {
		// Unmarshal DynamoDB item to dynamoItem
		dItem, err := unmarshalDynamoItem(item)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal dynamodb item: %w", err)
		}

		// Convert dynamoItem to domain entity
		entity, err := dynamoItemToEntity(dItem)
		if err != nil {
			return nil, fmt.Errorf("failed to convert item to entity: %w", err)
		}

		rates = append(rates, entity)
	}

	// Return empty slice (not nil) per interface contract
	return rates, nil
}

// Delete removes an exchange rate for a specific currency pair.
//
// This method:
// - Builds the partition key from base and target currencies
// - Uses DeleteItem operation to remove the item
// - Returns entity.ErrRateNotFound if the rate doesn't exist
// - Uses ReturnValues to check if item existed before deletion
//
// Context cancellation: Returns error if ctx is cancelled.
func (r *DynamoDBRepository) Delete(ctx context.Context, base, target entity.CurrencyCode) error {
	// Check context before starting operation
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Build partition key
	pk := buildPartitionKey(base, target)

	// Prepare DeleteItem input
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
		},
		ReturnValues: types.ReturnValueAllOld,
	}

	// Execute DeleteItem
	result, err := r.client.DeleteItem(ctx, input)
	if err != nil {
		return mapDynamoDBError(err, "delete item")
	}

	// Check if item existed (ReturnValues returns attributes of deleted item)
	if result.Attributes == nil {
		return entity.ErrRateNotFound
	}

	return nil
}

// GetStale retrieves a stale (expired) exchange rate for fallback scenarios.
//
// This method:
// - Is semantically the same as Get() for Phase 3
// - Repository doesn't filter by TTL - use cases handle expiration validation
// - Explicitly indicates the caller wants stale data for fallback purposes
// - Returns entity.ErrRateNotFound if no rate exists (even if expired)
//
// Note: The distinction from Get() is semantic. Both methods return rates
// regardless of TTL expiration. Use cases are responsible for checking expiration
// using entity.ExchangeRate.IsExpired() or entity.ExchangeRate.IsValid().
//
// Context cancellation: Returns error if ctx is cancelled.
func (r *DynamoDBRepository) GetStale(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
	// For Phase 3, GetStale is the same as Get
	// Repository doesn't filter by TTL - use cases handle expiration
	return r.Get(ctx, base, target)
}
