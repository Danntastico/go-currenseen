package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/misterfancybg/go-currenseen/internal/application/usecase"
	"github.com/misterfancybg/go-currenseen/internal/infrastructure/adapter/api"
	"github.com/misterfancybg/go-currenseen/internal/infrastructure/adapter/dynamodb"
	lambdaadapter "github.com/misterfancybg/go-currenseen/internal/infrastructure/adapter/lambda"
	"github.com/misterfancybg/go-currenseen/internal/infrastructure/config"
	"github.com/misterfancybg/go-currenseen/internal/infrastructure/middleware"
	"github.com/misterfancybg/go-currenseen/pkg/circuitbreaker"
)

var (
	// Global dependencies - initialized once during Lambda cold start
	deps *lambdaadapter.HandlerDependencies
)

// initDependencies initializes all dependencies for Lambda handlers.
//
// This function:
// - Creates DynamoDB client and repository
// - Creates HTTP client and API provider
// - Creates circuit breaker and wraps provider
// - Creates use cases with all dependencies
//
// Dependencies are initialized once during Lambda cold start and reused
// across invocations for better performance.
func initDependencies(ctx context.Context) error {
	// Load configuration
	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		return fmt.Errorf("TABLE_NAME environment variable is required")
	}

	// Cache TTL from environment (default: 1 hour)
	cacheTTL := 1 * time.Hour
	if ttlStr := os.Getenv("CACHE_TTL"); ttlStr != "" {
		if parsed, err := time.ParseDuration(ttlStr); err == nil && parsed > 0 {
			cacheTTL = parsed
		}
	}

	// 1. Initialize DynamoDB repository
	dynamoClient, err := config.NewDynamoDBClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create DynamoDB client: %w", err)
	}

	repository := dynamodb.NewDynamoDBRepository(dynamoClient, tableName)

	// 2. Initialize API provider with circuit breaker
	apiConfig := config.LoadAPIConfig()
	httpClient := api.NewHTTPClient()

	// Create base provider
	baseProvider := api.NewCurrencyAPIProvider(httpClient, apiConfig.BaseURL)

	// Create circuit breaker
	circuitBreakerConfig := config.LoadCircuitBreakerConfig()
	circuitBreaker, err := circuitbreaker.NewCircuitBreaker(circuitBreakerConfig)
	if err != nil {
		return fmt.Errorf("failed to create circuit breaker: %w", err)
	}

	// Wrap provider with circuit breaker
	provider := api.NewCircuitBreakerProvider(baseProvider, circuitBreaker)

	// 3. Initialize use cases
	getRateUseCase := usecase.NewGetExchangeRateUseCase(repository, provider, cacheTTL)
	getAllRatesUseCase := usecase.NewGetAllRatesUseCase(repository, provider, cacheTTL)
	healthCheckUseCase := usecase.NewHealthCheckUseCase(repository)

	// 4. Create handler dependencies
	deps = &lambdaadapter.HandlerDependencies{
		GetRateUseCase:     getRateUseCase,
		GetAllRatesUseCase: getAllRatesUseCase,
		HealthCheckUseCase: healthCheckUseCase,
	}

	return nil
}

// routeRequest routes API Gateway requests to the appropriate handler.
//
// This function:
// - Extracts path and method from the event
// - Routes to the appropriate handler based on path
// - Returns 404 for unknown routes
func routeRequest(ctx context.Context, event events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	path := event.Path
	method := event.HTTPMethod

	// Route based on path and method
	switch {
	case path == "/health" && method == "GET":
		return lambdaadapter.HealthHandler(ctx, event, deps)

	case strings.HasPrefix(path, "/rates/") && method == "GET":
		// Check if path has two segments (base/target) or one segment (base)
		// Path format: /rates/{base} or /rates/{base}/{target}
		pathParts := strings.Split(strings.TrimPrefix(path, "/rates/"), "/")

		if len(pathParts) == 2 {
			// Two segments: /rates/{base}/{target}
			return lambdaadapter.GetRateHandler(ctx, event, deps)
		} else if len(pathParts) == 1 && pathParts[0] != "" {
			// One segment: /rates/{base}
			return lambdaadapter.GetAllRatesHandler(ctx, event, deps)
		}
		// Fall through to 404
	}

	// Unknown route - return 404
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNotFound,
		Body:       fmt.Sprintf(`{"error":"Route not found: %s %s"}`, method, path),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// handler is the main Lambda handler function.
//
// This function:
// - Initializes dependencies on first invocation (cold start)
// - Routes requests to appropriate handlers
// - Handles errors appropriately
func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Initialize dependencies if not already initialized
	if deps == nil {
		if err := initDependencies(ctx); err != nil {
			// Return error response if initialization fails
			return middleware.ErrorResponse(fmt.Errorf("failed to initialize dependencies: %w", err)), nil
		}
	}

	// Route request to appropriate handler
	response := routeRequest(ctx, event)
	return response, nil
}

func main() {
	// Start Lambda runtime
	// The handler function will be called for each API Gateway event
	lambda.Start(handler)
}
