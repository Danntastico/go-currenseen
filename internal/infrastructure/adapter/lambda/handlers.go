package lambda

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/misterfancybg/go-currenseen/internal/application/dto"
	"github.com/misterfancybg/go-currenseen/internal/infrastructure/middleware"
)

// GetRateUseCase defines the interface for getting a single exchange rate.
// This interface enables dependency injection and makes handlers testable.
type GetRateUseCase interface {
	Execute(ctx context.Context, req dto.GetRateRequest) (dto.RateResponse, error)
}

// GetAllRatesUseCase defines the interface for getting all exchange rates for a base currency.
// This interface enables dependency injection and makes handlers testable.
type GetAllRatesUseCase interface {
	Execute(ctx context.Context, req dto.GetRatesRequest) (dto.RatesResponse, error)
}

// HealthCheckUseCase defines the interface for health checking the service.
// This interface enables dependency injection and makes handlers testable.
type HealthCheckUseCase interface {
	Execute(ctx context.Context, req dto.HealthCheckRequest) (dto.HealthCheckResponse, error)
}

// HandlerDependencies holds all dependencies needed by Lambda handlers.
// This struct enables dependency injection and makes handlers testable.
type HandlerDependencies struct {
	GetRateUseCase     GetRateUseCase
	GetAllRatesUseCase GetAllRatesUseCase
	HealthCheckUseCase HealthCheckUseCase
}

// GetRateHandler handles GET /rates/{base}/{target} requests.
//
// This handler:
// - Validates the request (path parameters, HTTP method)
// - Extracts base and target currency codes
// - Calls GetExchangeRateUseCase
// - Formats and returns the response
//
// Returns:
// - 200 OK with rate data on success
// - 400 Bad Request for invalid input
// - 404 Not Found if rate not found
// - 503 Service Unavailable if circuit breaker is open
// - 500 Internal Server Error for other errors
func GetRateHandler(ctx context.Context, event events.APIGatewayProxyRequest, deps *HandlerDependencies) events.APIGatewayProxyResponse {
	// Validate request
	base, target, err := middleware.ValidateGetRateRequest(event)
	if err != nil {
		return middleware.ErrorResponse(err)
	}

	// Create request DTO
	req := dto.GetRateRequest{
		Base:   base.String(),
		Target: target.String(),
	}

	// Call use case
	resp, err := deps.GetRateUseCase.Execute(ctx, req)
	if err != nil {
		return middleware.ErrorResponse(err)
	}

	// Return success response
	return middleware.SuccessResponse(200, resp)
}

// GetAllRatesHandler handles GET /rates/{base} requests.
//
// This handler:
// - Validates the request (path parameters, HTTP method)
// - Extracts base currency code
// - Calls GetAllRatesUseCase
// - Formats and returns the response
//
// Returns:
// - 200 OK with rates data on success
// - 400 Bad Request for invalid input
// - 503 Service Unavailable if circuit breaker is open
// - 500 Internal Server Error for other errors
func GetAllRatesHandler(ctx context.Context, event events.APIGatewayProxyRequest, deps *HandlerDependencies) events.APIGatewayProxyResponse {
	// Validate request
	base, err := middleware.ValidateGetRatesRequest(event)
	if err != nil {
		return middleware.ErrorResponse(err)
	}

	// Create request DTO
	req := dto.GetRatesRequest{
		Base: base.String(),
	}

	// Call use case
	resp, err := deps.GetAllRatesUseCase.Execute(ctx, req)
	if err != nil {
		return middleware.ErrorResponse(err)
	}

	// Return success response
	return middleware.SuccessResponse(200, resp)
}

// HealthHandler handles GET /health requests.
//
// This handler:
// - Validates the request (HTTP method)
// - Calls HealthCheckUseCase
// - Formats and returns the response
//
// Returns:
// - 200 OK if service is healthy
// - 503 Service Unavailable if service is unhealthy
func HealthHandler(ctx context.Context, event events.APIGatewayProxyRequest, deps *HandlerDependencies) events.APIGatewayProxyResponse {
	// Validate request
	if err := middleware.ValidateHealthRequest(event); err != nil {
		return middleware.ErrorResponse(err)
	}

	// Create request DTO
	req := dto.HealthCheckRequest{}

	// Call use case
	resp, err := deps.HealthCheckUseCase.Execute(ctx, req)
	if err != nil {
		return middleware.ErrorResponse(err)
	}

	// Determine status code based on health status
	statusCode := 200
	if resp.Status == "unhealthy" {
		statusCode = 503
	}

	// Return response
	return middleware.SuccessResponse(statusCode, resp)
}
