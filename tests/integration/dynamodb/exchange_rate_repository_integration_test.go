//go:build integration
// +build integration

package dynamodb

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
	dynamodbadapter "github.com/misterfancybg/go-currenseen/internal/infrastructure/adapter/dynamodb"
)

const (
	testTableName = "ExchangeRatesTest"
	gsiName       = "BaseCurrencyIndex"
)

var (
	testRepo *dynamodbadapter.DynamoDBRepository
	testCtx  context.Context
)

// setupTestTable creates the test DynamoDB table with the required schema.
func setupTestTable(ctx context.Context, client *dynamodb.Client) error {
	// Check if table already exists
	_, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(testTableName),
	})
	if err == nil {
		// Table exists, delete it first
		_, err = client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
			TableName: aws.String(testTableName),
		})
		if err != nil {
			return err
		}
		// Wait for table to be deleted
		waiter := dynamodb.NewTableNotExistsWaiter(client)
		err = waiter.Wait(ctx, &dynamodb.DescribeTableInput{
			TableName: aws.String(testTableName),
		}, 30*time.Second)
		if err != nil {
			return err
		}
	}

	// Create table
	_, err = client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(testTableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("Base"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("PK"),
				KeyType:       types.KeyTypeHash,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String(gsiName),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("Base"),
						KeyType:       types.KeyTypeHash,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		return err
	}

	// Wait for table to be active
	waiter := dynamodb.NewTableExistsWaiter(client)
	return waiter.Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(testTableName),
	}, 30*time.Second)
}

// teardownTestTable deletes the test table.
func teardownTestTable(ctx context.Context, client *dynamodb.Client) error {
	_, err := client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(testTableName),
	})
	return err
}

// setupIntegrationTest sets up the test environment.
func setupIntegrationTest(t *testing.T) {
	ctx := context.Background()

	// Load AWS config
	// For DynamoDB Local, set AWS_ENDPOINT_URL=http://localhost:8000
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to load AWS config: %v", err)
	}

	// Override endpoint for DynamoDB Local if specified
	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		cfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           endpoint,
					SigningRegion: region,
				}, nil
			},
		)
	}

	client := dynamodb.NewFromConfig(cfg)

	// Create test table
	if err := setupTestTable(ctx, client); err != nil {
		t.Fatalf("Failed to setup test table: %v", err)
	}

	// Create repository
	testRepo = dynamodbadapter.NewDynamoDBRepository(client, testTableName)
	testCtx = ctx
}

// teardownIntegrationTest cleans up the test environment.
func teardownIntegrationTest(t *testing.T) {
	if testRepo == nil {
		return
	}

	// Note: In a real scenario, you might want to delete the table
	// For now, we'll leave it for manual cleanup or reuse
	// Uncomment if you want automatic cleanup:
	// ctx := context.Background()
	// cfg, _ := config.LoadDefaultConfig(ctx)
	// client := dynamodb.NewFromConfig(cfg)
	// _ = teardownTestTable(ctx, client)
}

func TestMain(m *testing.M) {
	// Check if integration tests should run
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		// Skip integration tests unless explicitly enabled
		os.Exit(0)
	}

	code := m.Run()
	os.Exit(code)
}

func TestDynamoDBRepository_Get_Success(t *testing.T) {
	setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	// Create test rate
	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")
	rate, err := entity.NewExchangeRate(base, target, 0.85, time.Now().Add(-1*time.Hour), false)
	if err != nil {
		t.Fatalf("Failed to create test rate: %v", err)
	}

	// Save rate
	ttl := 1 * time.Hour
	if err := testRepo.Save(testCtx, rate, ttl); err != nil {
		t.Fatalf("Failed to save rate: %v", err)
	}

	// Get rate
	got, err := testRepo.Get(testCtx, base, target)
	if err != nil {
		t.Fatalf("Failed to get rate: %v", err)
	}

	// Verify
	if got.Base.String() != rate.Base.String() {
		t.Errorf("Base = %v, want %v", got.Base.String(), rate.Base.String())
	}
	if got.Target.String() != rate.Target.String() {
		t.Errorf("Target = %v, want %v", got.Target.String(), rate.Target.String())
	}
	if got.Rate != rate.Rate {
		t.Errorf("Rate = %v, want %v", got.Rate, rate.Rate)
	}
}

func TestDynamoDBRepository_Get_NotFound(t *testing.T) {
	setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("JPY")

	// Try to get non-existent rate
	_, err := testRepo.Get(testCtx, base, target)
	if err == nil {
		t.Fatal("Expected ErrRateNotFound, got nil")
	}
	if !errors.Is(err, entity.ErrRateNotFound) {
		t.Errorf("Error = %v, want ErrRateNotFound", err)
	}
}

func TestDynamoDBRepository_Save_Success(t *testing.T) {
	setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	base, _ := entity.NewCurrencyCode("GBP")
	target, _ := entity.NewCurrencyCode("USD")
	rate, err := entity.NewExchangeRate(base, target, 1.25, time.Now().Add(-30*time.Minute), false)
	if err != nil {
		t.Fatalf("Failed to create test rate: %v", err)
	}

	// Save rate
	ttl := 1 * time.Hour
	if err := testRepo.Save(testCtx, rate, ttl); err != nil {
		t.Fatalf("Failed to save rate: %v", err)
	}

	// Verify it was saved
	got, err := testRepo.Get(testCtx, base, target)
	if err != nil {
		t.Fatalf("Failed to get saved rate: %v", err)
	}
	if got.Rate != rate.Rate {
		t.Errorf("Rate = %v, want %v", got.Rate, rate.Rate)
	}
}

func TestDynamoDBRepository_Save_Update(t *testing.T) {
	setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	base, _ := entity.NewCurrencyCode("EUR")
	target, _ := entity.NewCurrencyCode("GBP")
	rate1, _ := entity.NewExchangeRate(base, target, 0.88, time.Now().Add(-1*time.Hour), false)
	rate2, _ := entity.NewExchangeRate(base, target, 0.90, time.Now().Add(-30*time.Minute), false)

	// Save first rate
	if err := testRepo.Save(testCtx, rate1, 1*time.Hour); err != nil {
		t.Fatalf("Failed to save first rate: %v", err)
	}

	// Update with second rate
	if err := testRepo.Save(testCtx, rate2, 1*time.Hour); err != nil {
		t.Fatalf("Failed to update rate: %v", err)
	}

	// Verify update
	got, err := testRepo.Get(testCtx, base, target)
	if err != nil {
		t.Fatalf("Failed to get updated rate: %v", err)
	}
	if got.Rate != rate2.Rate {
		t.Errorf("Rate = %v, want %v (updated rate)", got.Rate, rate2.Rate)
	}
}

func TestDynamoDBRepository_GetByBase_Success(t *testing.T) {
	setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	base, _ := entity.NewCurrencyCode("USD")
	target1, _ := entity.NewCurrencyCode("EUR")
	target2, _ := entity.NewCurrencyCode("GBP")
	target3, _ := entity.NewCurrencyCode("JPY")

	// Save multiple rates for same base
	rate1, _ := entity.NewExchangeRate(base, target1, 0.85, time.Now().Add(-1*time.Hour), false)
	rate2, _ := entity.NewExchangeRate(base, target2, 0.75, time.Now().Add(-1*time.Hour), false)
	rate3, _ := entity.NewExchangeRate(base, target3, 150.0, time.Now().Add(-1*time.Hour), false)

	if err := testRepo.Save(testCtx, rate1, 1*time.Hour); err != nil {
		t.Fatalf("Failed to save rate1: %v", err)
	}
	if err := testRepo.Save(testCtx, rate2, 1*time.Hour); err != nil {
		t.Fatalf("Failed to save rate2: %v", err)
	}
	if err := testRepo.Save(testCtx, rate3, 1*time.Hour); err != nil {
		t.Fatalf("Failed to save rate3: %v", err)
	}

	// Get all rates for base
	rates, err := testRepo.GetByBase(testCtx, base)
	if err != nil {
		t.Fatalf("Failed to get rates by base: %v", err)
	}

	// Verify we got all rates
	if len(rates) < 3 {
		t.Errorf("Got %d rates, want at least 3", len(rates))
	}

	// Verify all rates have correct base
	for _, rate := range rates {
		if rate.Base.String() != base.String() {
			t.Errorf("Rate has base %v, want %v", rate.Base.String(), base.String())
		}
	}
}

func TestDynamoDBRepository_GetByBase_Empty(t *testing.T) {
	setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	base, _ := entity.NewCurrencyCode("CAD")

	// Get rates for base with no rates
	rates, err := testRepo.GetByBase(testCtx, base)
	if err != nil {
		t.Fatalf("Failed to get rates by base: %v", err)
	}

	// Should return empty slice, not nil
	if rates == nil {
		t.Error("Got nil, want empty slice")
	}
	if len(rates) != 0 {
		t.Errorf("Got %d rates, want 0", len(rates))
	}
}

func TestDynamoDBRepository_Delete_Success(t *testing.T) {
	setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	base, _ := entity.NewCurrencyCode("AUD")
	target, _ := entity.NewCurrencyCode("NZD")
	rate, _ := entity.NewExchangeRate(base, target, 1.10, time.Now().Add(-1*time.Hour), false)

	// Save rate
	if err := testRepo.Save(testCtx, rate, 1*time.Hour); err != nil {
		t.Fatalf("Failed to save rate: %v", err)
	}

	// Delete rate
	if err := testRepo.Delete(testCtx, base, target); err != nil {
		t.Fatalf("Failed to delete rate: %v", err)
	}

	// Verify it's deleted
	_, err := testRepo.Get(testCtx, base, target)
	if !errors.Is(err, entity.ErrRateNotFound) {
		t.Errorf("Expected ErrRateNotFound after delete, got: %v", err)
	}
}

func TestDynamoDBRepository_Delete_NotFound(t *testing.T) {
	setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	base, _ := entity.NewCurrencyCode("CHF")
	target, _ := entity.NewCurrencyCode("SEK")

	// Try to delete non-existent rate
	err := testRepo.Delete(testCtx, base, target)
	if !errors.Is(err, entity.ErrRateNotFound) {
		t.Errorf("Error = %v, want ErrRateNotFound", err)
	}
}

func TestDynamoDBRepository_GetStale_Success(t *testing.T) {
	setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	base, _ := entity.NewCurrencyCode("MXN")
	target, _ := entity.NewCurrencyCode("BRL")
	rate, _ := entity.NewExchangeRate(base, target, 5.50, time.Now().Add(-2*time.Hour), true)

	// Save stale rate
	if err := testRepo.Save(testCtx, rate, 1*time.Hour); err != nil {
		t.Fatalf("Failed to save stale rate: %v", err)
	}

	// Get stale rate (should work the same as Get)
	got, err := testRepo.GetStale(testCtx, base, target)
	if err != nil {
		t.Fatalf("Failed to get stale rate: %v", err)
	}

	// Verify
	if got.Rate != rate.Rate {
		t.Errorf("Rate = %v, want %v", got.Rate, rate.Rate)
	}
	if !got.Stale {
		t.Error("Expected stale flag to be true")
	}
}

func TestDynamoDBRepository_ContextCancellation(t *testing.T) {
	setupIntegrationTest(t)
	defer teardownIntegrationTest(t)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")

	// All operations should respect context cancellation
	_, err := testRepo.Get(ctx, base, target)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Get() error = %v, want context.Canceled", err)
	}

	rate, _ := entity.NewExchangeRate(base, target, 0.85, time.Now(), false)
	err = testRepo.Save(ctx, rate, 1*time.Hour)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Save() error = %v, want context.Canceled", err)
	}

	_, err = testRepo.GetByBase(ctx, base)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("GetByBase() error = %v, want context.Canceled", err)
	}

	err = testRepo.Delete(ctx, base, target)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Delete() error = %v, want context.Canceled", err)
	}
}
