package api

import (
	"crypto/tls"
	"net/http"
	"os"
	"strconv"
	"time"
)

// NewHTTPClient creates a new HTTP client with secure defaults.
//
// Configuration:
// - Timeout: 10 seconds (prevents hanging requests)
// - TLS: Minimum TLS 1.2 (security requirement)
// - Certificate Verification: Enabled by default (InsecureSkipVerify: false)
//   - Can be disabled for local development by setting SKIP_TLS_VERIFY=true
//
// - Transport: HTTP/1.1 (compatibility)
//
// The client is safe for concurrent use by multiple goroutines.
//
// This implementation follows security best practices:
// - Enforces TLS 1.2 minimum (prevents weak encryption)
// - Verifies SSL certificates by default (prevents MITM attacks)
// - Sets reasonable timeout (prevents resource exhaustion)
//
// WARNING: Setting SKIP_TLS_VERIFY=true is ONLY for local development.
// NEVER use this in production as it disables certificate verification.
//
// Example usage:
//
//	client := NewHTTPClient()
//	resp, err := client.Get("https://api.example.com/data")
//	if err != nil {
//	    // Handle error
//	}
//	defer resp.Body.Close()
func NewHTTPClient() *http.Client {
	// Check if TLS verification should be skipped (local development only)
	skipVerify := false
	if skipVerifyStr := os.Getenv("SKIP_TLS_VERIFY"); skipVerifyStr != "" {
		if val, err := strconv.ParseBool(skipVerifyStr); err == nil {
			skipVerify = val
		}
	}

	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: skipVerify, // Can be disabled for local dev
			},
			// Disable HTTP/2 for compatibility (can be enabled if needed)
			ForceAttemptHTTP2: false,
		},
	}
}
