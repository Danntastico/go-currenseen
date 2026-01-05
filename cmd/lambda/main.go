package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/misterfancybg/go-currenseen/internal/application/usecase"
	"github.com/misterfancybg/go-currenseen/internal/infrastructure/adapter/api"
	"github.com/misterfancybg/go-currenseen/internal/infrastructure/adapter/dynamodb"
	lambdaadapter "github.com/misterfancybg/go-currenseen/internal/infrastructure/adapter/lambda"
	"github.com/misterfancybg/go-currenseen/internal/infrastructure/config"
	"github.com/misterfancybg/go-currenseen/internal/infrastructure/middleware"
	"github.com/misterfancybg/go-currenseen/pkg/circuitbreaker"
	"github.com/misterfancybg/go-currenseen/pkg/logger"
)

var (
	// Global dependencies - initialized once during Lambda cold start
	deps *lambdaadapter.HandlerDependencies
)

// initDependencies initializes all dependencies for Lambda handlers.
//
// This function:
// - Loads unified configuration
// - Creates logger
// - Creates DynamoDB client and repository
// - Creates HTTP client and API provider
// - Creates circuit breaker and wraps provider
// - Creates use cases with all dependencies
// - Optionally initializes Secrets Manager for API keys
//
// Dependencies are initialized once during Lambda cold start and reused
// across invocations for better performance.
func initDependencies(ctx context.Context) error {
	// Initialize logger first
	log := logger.NewFromEnv()
	log.Info("initializing Lambda dependencies")

	// Load unified configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error("failed to load configuration", "error", err.Error())
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 1. Initialize DynamoDB repository
	dynamoClient, err := config.NewDynamoDBClient(ctx)
	if err != nil {
		log.Error("failed to create DynamoDB client", "error", err.Error())
		return fmt.Errorf("failed to create DynamoDB client: %w", err)
	}

	repository := dynamodb.NewDynamoDBRepository(dynamoClient, cfg.DynamoDB.TableName)

	// 2. Initialize API provider with circuit breaker
	httpClient := api.NewHTTPClient()

	// Create base provider with logger
	baseProvider := api.NewCurrencyAPIProvider(httpClient, cfg.API.BaseURL, log)

	// Create circuit breaker
	circuitBreaker, err := circuitbreaker.NewCircuitBreaker(cfg.CircuitBreaker)
	if err != nil {
		log.Error("failed to create circuit breaker", "error", err.Error())
		return fmt.Errorf("failed to create circuit breaker: %w", err)
	}

	// Wrap provider with circuit breaker
	provider := api.NewCircuitBreakerProvider(baseProvider, circuitBreaker)

	// 3. Initialize use cases with logger
	getRateUseCase := usecase.NewGetExchangeRateUseCase(repository, provider, cfg.Cache.TTL, log)
	getAllRatesUseCase := usecase.NewGetAllRatesUseCase(repository, provider, cfg.Cache.TTL, log)
	healthCheckUseCase := usecase.NewHealthCheckUseCase(repository)

	// 4. Initialize security components
	var apiKeyAuthenticator *middleware.APIKeyAuthenticator
	var rateLimiter *middleware.RateLimiter

	// Initialize Secrets Manager if enabled
	var secretsManager config.SecretsManager
	if cfg.SecretsManager.Enabled {
		sm, err := config.NewAWSSecretsManager(ctx, cfg.SecretsManager.SecretName, cfg.SecretsManager.CacheTTL)
		if err != nil {
			log.Warn("failed to initialize Secrets Manager, authentication will be disabled", "error", err.Error())
		} else {
			secretsManager = sm
			log.Info("Secrets Manager initialized", "secret_name", cfg.SecretsManager.SecretName)
		}
	}

	// Initialize API key authenticator (enabled by default if Secrets Manager is configured)
	authEnabled := cfg.SecretsManager.Enabled && secretsManager != nil
	if authEnabled {
		apiKeyAuthenticator = middleware.NewAPIKeyAuthenticator(secretsManager, cfg, true)
		log.Info("API key authentication enabled")
	} else {
		log.Info("API key authentication disabled (for development)")
	}

	// Initialize rate limiter (enabled by default)
	rateLimitConfig := middleware.DefaultRateLimiterConfig()
	// Allow configuration via environment variables
	if envRateLimit := os.Getenv("RATE_LIMIT_REQUESTS_PER_MINUTE"); envRateLimit != "" {
		if parsed, err := strconv.Atoi(envRateLimit); err == nil && parsed > 0 {
			rateLimitConfig.RequestsPerMinute = parsed
		}
	}
	if envBurst := os.Getenv("RATE_LIMIT_BURST_SIZE"); envBurst != "" {
		if parsed, err := strconv.Atoi(envBurst); err == nil && parsed > 0 {
			rateLimitConfig.BurstSize = parsed
		}
	}
	if os.Getenv("RATE_LIMIT_ENABLED") == "false" {
		rateLimitConfig.Enabled = false
	}

	rateLimiter = middleware.NewRateLimiter(rateLimitConfig)
	if rateLimitConfig.Enabled {
		log.Info("Rate limiting enabled",
			"requests_per_minute", rateLimitConfig.RequestsPerMinute,
			"burst_size", rateLimitConfig.BurstSize,
		)
	} else {
		log.Info("Rate limiting disabled")
	}

	// 5. Create handler dependencies
	deps = &lambdaadapter.HandlerDependencies{
		GetRateUseCase:      getRateUseCase,
		GetAllRatesUseCase:  getAllRatesUseCase,
		HealthCheckUseCase:  healthCheckUseCase,
		Logger:              log,
		APIKeyAuthenticator: apiKeyAuthenticator,
		RateLimiter:         rateLimiter,
	}

	log.Info("Lambda dependencies initialized successfully")
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
