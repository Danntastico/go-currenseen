package dynamodb

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
)

// DynamoDBRepository implements the ExchangeRateRepository interface using AWS DynamoDB.
// This is an adapter in the Hexagonal Architecture pattern, connecting the domain layer
// to the AWS DynamoDB infrastructure.

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
