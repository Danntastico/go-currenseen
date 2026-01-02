package middleware

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/misterfancybg/go-currenseen/pkg/logger"
)

func TestGenerateRequestID(t *testing.T) {
	id1 := GenerateRequestID()
	id2 := GenerateRequestID()

	if id1 == "" {
		t.Error("GenerateRequestID() returned empty string")
	}

	if id2 == "" {
		t.Error("GenerateRequestID() returned empty string")
	}

	// IDs should be different (very high probability)
	if id1 == id2 {
		t.Error("GenerateRequestID() returned duplicate IDs")
	}

	// ID should be hex-encoded (even length, valid hex chars)
	if len(id1)%2 != 0 {
		t.Errorf("GenerateRequestID() returned ID with odd length: %d", len(id1))
	}
}

func TestExtractOrGenerateRequestID_FromContext(t *testing.T) {
	event := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			RequestID: "aws-request-123",
		},
	}

	requestID := ExtractOrGenerateRequestID(event)
	if requestID != "aws-request-123" {
		t.Errorf("ExtractOrGenerateRequestID() = %q, want %q", requestID, "aws-request-123")
	}
}

func TestExtractOrGenerateRequestID_FromHeader(t *testing.T) {
	event := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"X-Request-ID": "header-request-456",
		},
	}

	requestID := ExtractOrGenerateRequestID(event)
	if requestID != "header-request-456" {
		t.Errorf("ExtractOrGenerateRequestID() = %q, want %q", requestID, "header-request-456")
	}
}

func TestExtractOrGenerateRequestID_FromHeaderLowercase(t *testing.T) {
	event := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"x-request-id": "lowercase-header-789",
		},
	}

	requestID := ExtractOrGenerateRequestID(event)
	if requestID != "lowercase-header-789" {
		t.Errorf("ExtractOrGenerateRequestID() = %q, want %q", requestID, "lowercase-header-789")
	}
}

func TestExtractOrGenerateRequestID_Generated(t *testing.T) {
	event := events.APIGatewayProxyRequest{}

	requestID := ExtractOrGenerateRequestID(event)
	if requestID == "" {
		t.Error("ExtractOrGenerateRequestID() returned empty string")
	}
}

func TestWithRequestID(t *testing.T) {
	event := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			RequestID: "test-request-123",
		},
	}

	ctx := context.Background()
	ctx = WithRequestID(ctx, event)

	requestID := logger.GetRequestID(ctx)
	if requestID != "test-request-123" {
		t.Errorf("GetRequestID() = %q, want %q", requestID, "test-request-123")
	}
}

func TestWithRequestID_Priority(t *testing.T) {
	// Test that context request ID takes priority over header
	event := events.APIGatewayProxyRequest{
		RequestContext: events.APIGatewayProxyRequestContext{
			RequestID: "context-id",
		},
		Headers: map[string]string{
			"X-Request-ID": "header-id",
		},
	}

	ctx := context.Background()
	ctx = WithRequestID(ctx, event)

	requestID := logger.GetRequestID(ctx)
	if requestID != "context-id" {
		t.Errorf("GetRequestID() = %q, want %q (context should take priority)", requestID, "context-id")
	}
}
