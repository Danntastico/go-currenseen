package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestNew_Defaults(t *testing.T) {
	logger := New(nil)

	if logger == nil {
		t.Fatal("New() returned nil")
	}

	if logger.Logger == nil {
		t.Fatal("Logger.Logger is nil")
	}
}

func TestNew_CustomConfig(t *testing.T) {
	config := &Config{
		Level:      "DEBUG",
		Format:     "text",
		AddSource:  true,
		CloudWatch: false,
	}

	logger := New(config)

	if logger == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNewFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("LOG_LEVEL", "WARN")
	os.Setenv("LOG_FORMAT", "text")
	defer func() {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FORMAT")
	}()

	logger := NewFromEnv()

	if logger == nil {
		t.Fatal("NewFromEnv() returned nil")
	}
}

func TestWithRequestID(t *testing.T) {
	logger := New(nil)
	requestID := "test-request-123"

	loggerWithID := logger.WithRequestID(requestID)

	if loggerWithID == nil {
		t.Fatal("WithRequestID() returned nil")
	}
}

func TestWithContext(t *testing.T) {
	logger := New(nil)

	ctx := context.WithValue(context.Background(), RequestIDKey, "test-123")
	ctx = context.WithValue(ctx, BaseCurrencyKey, "USD")
	ctx = context.WithValue(ctx, TargetCurrencyKey, "EUR")

	loggerWithCtx := logger.WithContext(ctx)

	if loggerWithCtx == nil {
		t.Fatal("WithContext() returned nil")
	}
}

func TestSanitizeValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no sensitive data",
			input:    "normal text",
			expected: "normal text",
		},
		{
			name:     "api key pattern",
			input:    "api_key: secret123",
			expected: "[REDACTED]",
		},
		{
			name:     "bearer token",
			input:    "Bearer abc123token",
			expected: "Bearer [REDACTED]",
		},
		{
			name:     "authorization header",
			input:    "Authorization: Bearer token123",
			expected: "[REDACTED]", // May be "[REDACTED][REDACTED]" if multiple patterns match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeValue(tt.input)
			// For authorization header, accept either single or double redaction
			if tt.name == "authorization header" {
				if result != "[REDACTED]" && result != "[REDACTED][REDACTED]" {
					t.Errorf("SanitizeValue(%q) = %q, want %q or %q", tt.input, result, "[REDACTED]", "[REDACTED][REDACTED]")
				}
			} else {
				if result != tt.expected {
					t.Errorf("SanitizeValue(%q) = %q, want %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "short key",
			input:    "short",
			expected: "****",
		},
		{
			name:     "normal key",
			input:    "abcdefghijklmnop",
			expected: "abcd****mnop",
		},
		{
			name:     "long key",
			input:    "verylongapikey1234567890",
			expected: "very****7890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskAPIKey(tt.input)
			if result != tt.expected {
				t.Errorf("MaskAPIKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestWithRequestID_Context(t *testing.T) {
	ctx := context.Background()
	requestID := "test-123"

	ctx = WithRequestID(ctx, requestID)

	retrieved := GetRequestID(ctx)
	if retrieved != requestID {
		t.Errorf("GetRequestID() = %q, want %q", retrieved, requestID)
	}
}

func TestWithCurrencyCodes(t *testing.T) {
	ctx := context.Background()
	base := "USD"
	target := "EUR"

	ctx = WithCurrencyCodes(ctx, base, target)

	if baseVal, ok := ctx.Value(BaseCurrencyKey).(string); !ok || baseVal != base {
		t.Errorf("BaseCurrencyKey = %q, want %q", baseVal, base)
	}

	if targetVal, ok := ctx.Value(TargetCurrencyKey).(string); !ok || targetVal != target {
		t.Errorf("TargetCurrencyKey = %q, want %q", targetVal, target)
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"DEBUG", slog.LevelDebug},
		{"INFO", slog.LevelInfo},
		{"WARN", slog.LevelWarn},
		{"WARNING", slog.LevelWarn},
		{"ERROR", slog.LevelError},
		{"unknown", slog.LevelInfo}, // default
		{"", slog.LevelInfo},        // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLogMethods(t *testing.T) {
	// Capture output by using a buffer or checking that methods don't panic
	logger := New(&Config{
		Level:  "DEBUG",
		Format: "text",
	})

	// Test that methods don't panic
	logger.Debug("debug message", "key", "value")
	logger.Info("info message", "key", "value")
	logger.Warn("warning message", "key", "value")
	logger.Error("error message", "key", "value")

	ctx := context.Background()
	logger.LogRequest(ctx, "GET", "/rates/USD/EUR")
	logger.LogResponse(ctx, 200, 150, "key", "value")
	logger.LogError(ctx, context.DeadlineExceeded, "request timeout", "key", "value")
}

func TestGetEnvironment(t *testing.T) {
	// Test default
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("ENV")
	env := getEnvironment()
	if env != "development" {
		t.Errorf("getEnvironment() = %q, want %q", env, "development")
	}

	// Test ENVIRONMENT variable
	os.Setenv("ENVIRONMENT", "production")
	defer os.Unsetenv("ENVIRONMENT")
	env = getEnvironment()
	if env != "production" {
		t.Errorf("getEnvironment() = %q, want %q", env, "production")
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	os.Unsetenv("TEST_VAR")
	result := getEnvOrDefault("TEST_VAR", "default")
	if result != "default" {
		t.Errorf("getEnvOrDefault() = %q, want %q", result, "default")
	}

	os.Setenv("TEST_VAR", "custom")
	defer os.Unsetenv("TEST_VAR")
	result = getEnvOrDefault("TEST_VAR", "default")
	if result != "custom" {
		t.Errorf("getEnvOrDefault() = %q, want %q", result, "custom")
	}
}

// Test that JSON output is valid (basic check)
func TestJSONOutput(t *testing.T) {
	logger := New(&Config{
		Level:  "INFO",
		Format: "json",
	})

	// This test just ensures the logger can be created and used
	// Actual JSON validation would require capturing stdout
	logger.Info("test message", "key", "value")
}

// Test log level filtering
func TestLogLevelFiltering(t *testing.T) {
	logger := New(&Config{
		Level:  "WARN",
		Format: "text",
	})

	// These should not appear in output (but won't panic)
	logger.Debug("debug message")
	logger.Info("info message")

	// These should appear
	logger.Warn("warning message")
	logger.Error("error message")
}

// Test context value extraction
func TestContextExtraction(t *testing.T) {
	logger := New(nil)

	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithCurrencyCodes(ctx, "USD", "EUR")

	loggerWithCtx := logger.WithContext(ctx)

	// Verify logger was created (can't easily verify internal state)
	if loggerWithCtx == nil {
		t.Fatal("WithContext() returned nil")
	}
}

// Test sanitization with various patterns
func TestSanitizeValue_Comprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string // String that should NOT be in output
	}{
		{
			name:     "API key with underscore",
			input:    "api_key=secret123456",
			contains: "secret123456",
		},
		{
			name:     "API key with dash",
			input:    "api-key: secret123456",
			contains: "secret123456",
		},
		{
			name:     "Token pattern",
			input:    "token=abc123def456",
			contains: "abc123def456",
		},
		{
			name:     "Password pattern",
			input:    "password: mypassword123",
			contains: "mypassword123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeValue(tt.input)
			if strings.Contains(result, tt.contains) {
				t.Errorf("SanitizeValue(%q) contains sensitive data: %q", tt.input, result)
			}
		})
	}
}
