package lambda

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/misterfancybg/go-currenseen/internal/application/dto"
	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
)

// mockGetRateUseCase is a mock implementation of GetExchangeRateUseCase for testing.
type mockGetRateUseCase struct {
	executeFunc func(ctx context.Context, req dto.GetRateRequest) (dto.RateResponse, error)
}

func (m *mockGetRateUseCase) Execute(ctx context.Context, req dto.GetRateRequest) (dto.RateResponse, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, req)
	}
	return dto.RateResponse{}, errors.New("not implemented")
}

// mockGetAllRatesUseCase is a mock implementation of GetAllRatesUseCase for testing.
type mockGetAllRatesUseCase struct {
	executeFunc func(ctx context.Context, req dto.GetRatesRequest) (dto.RatesResponse, error)
}

func (m *mockGetAllRatesUseCase) Execute(ctx context.Context, req dto.GetRatesRequest) (dto.RatesResponse, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, req)
	}
	return dto.RatesResponse{}, errors.New("not implemented")
}

// mockHealthCheckUseCase is a mock implementation of HealthCheckUseCase for testing.
type mockHealthCheckUseCase struct {
	executeFunc func(ctx context.Context, req dto.HealthCheckRequest) (dto.HealthCheckResponse, error)
}

func (m *mockHealthCheckUseCase) Execute(ctx context.Context, req dto.HealthCheckRequest) (dto.HealthCheckResponse, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, req)
	}
	return dto.HealthCheckResponse{}, errors.New("not implemented")
}

func TestGetRateHandler_Success(t *testing.T) {
	ctx := context.Background()
	event := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/rates/USD/EUR",
		PathParameters: map[string]string{
			"base":   "USD",
			"target": "EUR",
		},
	}

	expectedResponse := dto.RateResponse{
		Base:      "USD",
		Target:    "EUR",
		Rate:      0.85,
		Timestamp: time.Now(),
		Stale:     false,
	}

	deps := &HandlerDependencies{
		GetRateUseCase: &mockGetRateUseCase{
			executeFunc: func(ctx context.Context, req dto.GetRateRequest) (dto.RateResponse, error) {
				if req.Base != "USD" || req.Target != "EUR" {
					t.Errorf("unexpected request: base=%s, target=%s", req.Base, req.Target)
				}
				return expectedResponse, nil
			},
		},
	}

	resp := GetRateHandler(ctx, event, deps)

	if resp.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", resp.StatusCode)
	}

	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", resp.Headers["Content-Type"])
	}
}

func TestGetRateHandler_InvalidCurrencyCode(t *testing.T) {
	ctx := context.Background()
	event := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/rates/XX/EUR",
		PathParameters: map[string]string{
			"base":   "XX",
			"target": "EUR",
		},
	}

	deps := &HandlerDependencies{
		GetRateUseCase: &mockGetRateUseCase{},
	}

	resp := GetRateHandler(ctx, event, deps)

	if resp.StatusCode != 400 {
		t.Errorf("expected status code 400, got %d", resp.StatusCode)
	}
}

func TestGetRateHandler_MissingPathParameter(t *testing.T) {
	ctx := context.Background()
	event := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/rates/USD",
		PathParameters: map[string]string{
			"base": "USD",
		},
	}

	deps := &HandlerDependencies{
		GetRateUseCase: &mockGetRateUseCase{},
	}

	resp := GetRateHandler(ctx, event, deps)

	if resp.StatusCode != 400 {
		t.Errorf("expected status code 400, got %d", resp.StatusCode)
	}
}

func TestGetRateHandler_UseCaseError(t *testing.T) {
	ctx := context.Background()
	event := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/rates/USD/EUR",
		PathParameters: map[string]string{
			"base":   "USD",
			"target": "EUR",
		},
	}

	deps := &HandlerDependencies{
		GetRateUseCase: &mockGetRateUseCase{
			executeFunc: func(ctx context.Context, req dto.GetRateRequest) (dto.RateResponse, error) {
				return dto.RateResponse{}, entity.ErrRateNotFound
			},
		},
	}

	resp := GetRateHandler(ctx, event, deps)

	if resp.StatusCode != 404 {
		t.Errorf("expected status code 404, got %d", resp.StatusCode)
	}
}

func TestGetAllRatesHandler_Success(t *testing.T) {
	ctx := context.Background()
	event := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/rates/USD",
		PathParameters: map[string]string{
			"base": "USD",
		},
	}

	expectedResponse := dto.RatesResponse{
		Base:      "USD",
		Rates:     make(map[string]dto.RateResponse),
		Timestamp: time.Now(),
		Stale:     false,
	}

	deps := &HandlerDependencies{
		GetAllRatesUseCase: &mockGetAllRatesUseCase{
			executeFunc: func(ctx context.Context, req dto.GetRatesRequest) (dto.RatesResponse, error) {
				if req.Base != "USD" {
					t.Errorf("unexpected request: base=%s", req.Base)
				}
				return expectedResponse, nil
			},
		},
	}

	resp := GetAllRatesHandler(ctx, event, deps)

	if resp.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", resp.StatusCode)
	}
}

func TestGetAllRatesHandler_InvalidCurrencyCode(t *testing.T) {
	ctx := context.Background()
	event := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/rates/XX",
		PathParameters: map[string]string{
			"base": "XX",
		},
	}

	deps := &HandlerDependencies{
		GetAllRatesUseCase: &mockGetAllRatesUseCase{},
	}

	resp := GetAllRatesHandler(ctx, event, deps)

	if resp.StatusCode != 400 {
		t.Errorf("expected status code 400, got %d", resp.StatusCode)
	}
}

func TestHealthHandler_Success(t *testing.T) {
	ctx := context.Background()
	event := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/health",
	}

	expectedResponse := dto.HealthCheckResponse{
		Status:    "healthy",
		Checks:    map[string]string{"lambda": "healthy", "dynamodb": "healthy"},
		Timestamp: time.Now(),
	}

	deps := &HandlerDependencies{
		HealthCheckUseCase: &mockHealthCheckUseCase{
			executeFunc: func(ctx context.Context, req dto.HealthCheckRequest) (dto.HealthCheckResponse, error) {
				return expectedResponse, nil
			},
		},
	}

	resp := HealthHandler(ctx, event, deps)

	if resp.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", resp.StatusCode)
	}
}

func TestHealthHandler_Unhealthy(t *testing.T) {
	ctx := context.Background()
	event := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/health",
	}

	expectedResponse := dto.HealthCheckResponse{
		Status:    "unhealthy",
		Checks:    map[string]string{"lambda": "healthy", "dynamodb": "unhealthy"},
		Timestamp: time.Now(),
	}

	deps := &HandlerDependencies{
		HealthCheckUseCase: &mockHealthCheckUseCase{
			executeFunc: func(ctx context.Context, req dto.HealthCheckRequest) (dto.HealthCheckResponse, error) {
				return expectedResponse, nil
			},
		},
	}

	resp := HealthHandler(ctx, event, deps)

	if resp.StatusCode != 503 {
		t.Errorf("expected status code 503, got %d", resp.StatusCode)
	}
}
