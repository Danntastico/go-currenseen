package dynamodb

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
)

// Helper to create a test exchange rate entity
func createTestExchangeRate() (*entity.ExchangeRate, error) {
	base, err := entity.NewCurrencyCode("USD")
	if err != nil {
		return nil, err
	}
	target, err := entity.NewCurrencyCode("EUR")
	if err != nil {
		return nil, err
	}
	return entity.NewExchangeRate(base, target, 0.85, time.Now().Add(-1*time.Hour), false)
}

func TestEntityToDynamoItem(t *testing.T) {
	rate, err := createTestExchangeRate()
	if err != nil {
		t.Fatalf("Failed to create test exchange rate: %v", err)
	}

	tests := []struct {
		name     string
		rate     *entity.ExchangeRate
		ttl      time.Duration
		wantErr  bool
		checkTTL func(*dynamoItem) bool
	}{
		{
			name:    "valid rate with TTL",
			rate:    rate,
			ttl:     1 * time.Hour,
			wantErr: false,
			checkTTL: func(item *dynamoItem) bool {
				return item.TTL != nil && *item.TTL > 0
			},
		},
		{
			name:    "valid rate without TTL",
			rate:    rate,
			ttl:     0,
			wantErr: false,
			checkTTL: func(item *dynamoItem) bool {
				return item.TTL == nil
			},
		},
		{
			name:    "nil rate",
			rate:    nil,
			ttl:     1 * time.Hour,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := entityToDynamoItem(tt.rate, tt.ttl)
			if (err != nil) != tt.wantErr {
				t.Errorf("entityToDynamoItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && item != nil {
				if tt.checkTTL != nil && !tt.checkTTL(item) {
					t.Errorf("TTL check failed")
				}
				if item.PK != "RATE#USD#EUR" {
					t.Errorf("PK = %v, want RATE#USD#EUR", item.PK)
				}
				if item.Base != "USD" {
					t.Errorf("Base = %v, want USD", item.Base)
				}
				if item.Target != "EUR" {
					t.Errorf("Target = %v, want EUR", item.Target)
				}
			}
		})
	}
}

func TestDynamoItemToEntity(t *testing.T) {
	rate, err := createTestExchangeRate()
	if err != nil {
		t.Fatalf("Failed to create test exchange rate: %v", err)
	}

	item, err := entityToDynamoItem(rate, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create dynamo item: %v", err)
	}

	tests := []struct {
		name    string
		item    *dynamoItem
		wantErr bool
	}{
		{
			name:    "valid item",
			item:    item,
			wantErr: false,
		},
		{
			name:    "nil item",
			item:    nil,
			wantErr: true,
		},
		{
			name: "invalid base currency",
			item: &dynamoItem{
				PK:        "RATE#INVALID#EUR",
				Base:      "INVALID",
				Target:    "EUR",
				Rate:      0.85,
				Timestamp: time.Now().Unix(),
				Stale:     false,
			},
			wantErr: true,
		},
		{
			name: "invalid target currency",
			item: &dynamoItem{
				PK:        "RATE#USD#INVALID",
				Base:      "USD",
				Target:    "INVALID",
				Rate:      0.85,
				Timestamp: time.Now().Unix(),
				Stale:     false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, err := dynamoItemToEntity(tt.item)
			if (err != nil) != tt.wantErr {
				t.Errorf("dynamoItemToEntity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && entity != nil {
				if entity.Base.String() != "USD" {
					t.Errorf("Base = %v, want USD", entity.Base.String())
				}
				if entity.Target.String() != "EUR" {
					t.Errorf("Target = %v, want EUR", entity.Target.String())
				}
				if entity.Rate != 0.85 {
					t.Errorf("Rate = %v, want 0.85", entity.Rate)
				}
			}
		})
	}
}

func TestBuildPartitionKey(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")

	tests := []struct {
		name   string
		base   entity.CurrencyCode
		target entity.CurrencyCode
		want   string
	}{
		{
			name:   "USD to EUR",
			base:   base,
			target: target,
			want:   "RATE#USD#EUR",
		},
		{
			name:   "GBP to JPY",
			base:   entity.CurrencyCode("GBP"),
			target: entity.CurrencyCode("JPY"),
			want:   "RATE#GBP#JPY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPartitionKey(tt.base, tt.target)
			if got != tt.want {
				t.Errorf("buildPartitionKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshalUnmarshalDynamoItem(t *testing.T) {
	rate, err := createTestExchangeRate()
	if err != nil {
		t.Fatalf("Failed to create test exchange rate: %v", err)
	}

	item, err := entityToDynamoItem(rate, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create dynamo item: %v", err)
	}

	// Marshal
	av, err := marshalDynamoItem(item)
	if err != nil {
		t.Fatalf("marshalDynamoItem() error = %v", err)
	}

	// Unmarshal
	unmarshaledItem, err := unmarshalDynamoItem(av)
	if err != nil {
		t.Fatalf("unmarshalDynamoItem() error = %v", err)
	}

	// Verify
	if unmarshaledItem.PK != item.PK {
		t.Errorf("PK = %v, want %v", unmarshaledItem.PK, item.PK)
	}
	if unmarshaledItem.Base != item.Base {
		t.Errorf("Base = %v, want %v", unmarshaledItem.Base, item.Base)
	}
	if unmarshaledItem.Target != item.Target {
		t.Errorf("Target = %v, want %v", unmarshaledItem.Target, item.Target)
	}
	if unmarshaledItem.Rate != item.Rate {
		t.Errorf("Rate = %v, want %v", unmarshaledItem.Rate, item.Rate)
	}
}

func TestMapDynamoDBError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		operation string
		wantNil   bool
		checkErr  func(error) bool
	}{
		{
			name:      "nil error",
			err:       nil,
			operation: "test",
			wantNil:   true,
		},
		{
			name:      "context canceled",
			err:       context.Canceled,
			operation: "test",
			wantNil:   false,
			checkErr: func(err error) bool {
				return errors.Is(err, context.Canceled)
			},
		},
		{
			name:      "context deadline exceeded",
			err:       context.DeadlineExceeded,
			operation: "test",
			wantNil:   false,
			checkErr: func(err error) bool {
				return errors.Is(err, context.DeadlineExceeded)
			},
		},
		{
			name:      "resource not found",
			err:       &types.ResourceNotFoundException{Message: aws.String("Table not found")},
			operation: "get item",
			wantNil:   false,
			checkErr: func(err error) bool {
				return err != nil && err.Error() != ""
			},
		},
		{
			name:      "generic error",
			err:       errors.New("some error"),
			operation: "put item",
			wantNil:   false,
			checkErr: func(err error) bool {
				return err != nil && err.Error() != ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapDynamoDBError(tt.err, tt.operation)
			if (got == nil) != tt.wantNil {
				t.Errorf("mapDynamoDBError() = %v, want nil=%v", got, tt.wantNil)
				return
			}
			if !tt.wantNil && tt.checkErr != nil {
				if !tt.checkErr(got) {
					t.Errorf("mapDynamoDBError() check failed for error: %v", got)
				}
			}
		})
	}
}

func TestNewDynamoDBRepository(t *testing.T) {
	cfg := aws.Config{}
	client := dynamodb.NewFromConfig(cfg)
	tableName := "TestTable"

	repo := NewDynamoDBRepository(client, tableName)

	if repo == nil {
		t.Fatal("NewDynamoDBRepository() returned nil")
	}
	if repo.tableName != tableName {
		t.Errorf("tableName = %v, want %v", repo.tableName, tableName)
	}
	if repo.client == nil {
		t.Error("client is nil")
	}
}

// Note: Full integration tests for Get, Save, GetByBase, Delete, and GetStale
// would require either:
// 1. A real DynamoDB instance (local or test table)
// 2. A more sophisticated mocking library (like testify/mock)
// 3. An interface wrapper around the DynamoDB client
//
// The tests above cover the core conversion and helper functions.
// For full method tests, see integration tests or use a mocking framework.
