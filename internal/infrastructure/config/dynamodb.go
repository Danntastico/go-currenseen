package config

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// NewDynamoDBClient creates a new DynamoDB client with default configuration.
//
// In production (AWS Lambda), this will automatically use IAM role credentials.
// For local development, it uses credentials from:
//   - Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION)
//   - AWS credentials file (~/.aws/credentials)
//   - AWS config file (~/.aws/config)
//
// The client is safe for concurrent use by multiple goroutines.
//
// Example usage:
//
//	ctx := context.Background()
//	client, err := NewDynamoDBClient(ctx)
//	if err != nil {
//	    log.Fatalf("failed to create DynamoDB client: %v", err)
//	}
func NewDynamoDBClient(ctx context.Context) (*dynamodb.Client, error) {
	// Load default AWS configuration
	// This automatically handles credentials from:
	// 1. Environment variables
	// 2. AWS credentials file
	// 3. IAM role (when running on AWS)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create DynamoDB client from configuration
	return dynamodb.NewFromConfig(cfg), nil
}

// DynamoDBConfig holds configuration for DynamoDB repository.
// This struct can be used to pass both client and table name together.
type DynamoDBConfig struct {
	TableName string
	Client    *dynamodb.Client
}
