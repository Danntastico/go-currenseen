package logger

import (
	"context"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

// Context keys for logger values
type contextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey contextKey = "request_id"
	// BaseCurrencyKey is the context key for base currency
	BaseCurrencyKey contextKey = "base_currency"
	// TargetCurrencyKey is the context key for target currency
	TargetCurrencyKey contextKey = "target_currency"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
}

// Config holds logger configuration
type Config struct {
	Level      string // DEBUG, INFO, WARN, ERROR (default: INFO)
	Format     string // json or text (default: json)
	AddSource  bool   // Include source file/line in logs (default: false)
	CloudWatch bool   // Optimize for CloudWatch (default: true)
}

// New creates a new logger with the given configuration.
// If config is nil, uses defaults optimized for CloudWatch.
func New(config *Config) *Logger {
	if config == nil {
		config = &Config{
			Level:      "INFO",
			Format:     "json",
			AddSource:  false,
			CloudWatch: true,
		}
	}

	// Parse log level
	level := parseLogLevel(config.Level)

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: config.AddSource,
	}

	// Create handler based on format
	var handler slog.Handler
	if config.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		// JSON format (default, CloudWatch-friendly)
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	// Create logger
	logger := slog.New(handler)

	// Add default attributes for CloudWatch
	if config.CloudWatch {
		logger = logger.With(
			"service", "currency-exchange-rate",
			"environment", getEnvironment(),
		)
	}

	return &Logger{Logger: logger}
}

// NewFromEnv creates a logger from environment variables.
// Environment variables:
// - LOG_LEVEL: DEBUG, INFO, WARN, ERROR (default: INFO)
// - LOG_FORMAT: json or text (default: json)
// - LOG_SOURCE: true/false to include source file/line (default: false)
func NewFromEnv() *Logger {
	config := &Config{
		Level:      getEnvOrDefault("LOG_LEVEL", "INFO"),
		Format:     getEnvOrDefault("LOG_FORMAT", "json"),
		AddSource:  getEnvOrDefault("LOG_SOURCE", "false") == "true",
		CloudWatch: true,
	}

	return New(config)
}

// WithRequestID adds request ID to the logger context
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{Logger: l.Logger.With("request_id", requestID)}
}

// WithContext creates a logger with values from context
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.Logger

	// Extract request ID from context
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		logger = logger.With("request_id", requestID)
	}

	// Extract base currency from context
	if base, ok := ctx.Value(BaseCurrencyKey).(string); ok && base != "" {
		logger = logger.With("base_currency", base)
	}

	// Extract target currency from context
	if target, ok := ctx.Value(TargetCurrencyKey).(string); ok && target != "" {
		logger = logger.With("target_currency", target)
	}

	return &Logger{Logger: logger}
}

// Debug logs a debug message with optional key-value pairs
func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

// Info logs an info message with optional key-value pairs
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// Warn logs a warning message with optional key-value pairs
func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

// Error logs an error message with optional key-value pairs
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

// LogRequest logs an incoming HTTP request
func (l *Logger) LogRequest(ctx context.Context, method, path string, args ...any) {
	attrs := []any{"method", method, "path", path}
	attrs = append(attrs, args...)
	l.WithContext(ctx).Info("incoming request", attrs...)
}

// LogResponse logs an HTTP response
func (l *Logger) LogResponse(ctx context.Context, statusCode int, durationMs int64, args ...any) {
	attrs := []any{"status_code", statusCode, "duration_ms", durationMs}
	attrs = append(attrs, args...)
	l.WithContext(ctx).Info("request completed", attrs...)
}

// LogError logs an error with context
func (l *Logger) LogError(ctx context.Context, err error, msg string, args ...any) {
	attrs := []any{"error", err.Error()}
	attrs = append(attrs, args...)
	l.WithContext(ctx).Error(msg, attrs...)
}

// SanitizeValue sanitizes a value to prevent logging sensitive data
func SanitizeValue(value string) string {
	if value == "" {
		return value
	}

	// Patterns to redact
	patterns := []struct {
		pattern *regexp.Regexp
		replace string
	}{
		{regexp.MustCompile(`(?i)(api[_-]?key|token|password|secret)\s*[:=]\s*[\w-]+`), "[REDACTED]"},
		{regexp.MustCompile(`Bearer\s+[\w-]+`), "Bearer [REDACTED]"},
		{regexp.MustCompile(`(?i)(authorization|auth)\s*:\s*[\w\s-]+`), "[REDACTED]"},
	}

	sanitized := value
	for _, p := range patterns {
		sanitized = p.pattern.ReplaceAllString(sanitized, p.replace)
	}

	return sanitized
}

// MaskAPIKey masks an API key, showing only first 4 and last 4 characters
func MaskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) < 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// Helper functions

func parseLogLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func getEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = os.Getenv("ENV")
	}
	if env == "" {
		env = "development"
	}
	return env
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// WithCurrencyCodes adds currency codes to context
func WithCurrencyCodes(ctx context.Context, base, target string) context.Context {
	ctx = context.WithValue(ctx, BaseCurrencyKey, base)
	ctx = context.WithValue(ctx, TargetCurrencyKey, target)
	return ctx
}
