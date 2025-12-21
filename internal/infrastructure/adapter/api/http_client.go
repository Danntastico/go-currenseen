package api

import (
	"crypto/tls"
	"net/http"
	"time"
)

// NewHTTPClient creates a new HTTP client with secure defaults.
//
// Configuration:
// - Timeout: 10 seconds (prevents hanging requests)
// - TLS: Minimum TLS 1.2 (security requirement)
// - Certificate Verification: Enabled (InsecureSkipVerify: false)
// - Transport: HTTP/1.1 (compatibility)
//
// The client is safe for concurrent use by multiple goroutines.
//
// This implementation follows security best practices:
// - Enforces TLS 1.2 minimum (prevents weak encryption)
// - Verifies SSL certificates (prevents MITM attacks)
// - Sets reasonable timeout (prevents resource exhaustion)
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
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: false, // Always verify certificates
			},
			// Disable HTTP/2 for compatibility (can be enabled if needed)
			ForceAttemptHTTP2: false,
		},
	}
}
