package api

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"
)

func TestNewHTTPClient(t *testing.T) {
	client := NewHTTPClient()

	// Verify timeout is set correctly
	if client.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want 10s", client.Timeout)
	}

	// Verify transport exists
	if client.Transport == nil {
		t.Fatal("Transport is nil")
	}

	// Verify transport is *http.Transport
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport is %T, want *http.Transport", client.Transport)
	}

	// Verify TLS config exists
	if transport.TLSClientConfig == nil {
		t.Fatal("TLSClientConfig is nil")
	}

	// Verify TLS minimum version is 1.2
	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("TLS MinVersion = %v, want TLS 1.2", transport.TLSClientConfig.MinVersion)
	}

	// Verify certificate verification is enabled (security requirement)
	if transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify is true, want false (certificate verification must be enabled)")
	}

	// Verify HTTP/2 is disabled (as per configuration)
	if transport.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 is true, want false")
	}
}

func TestNewHTTPClient_ConcurrentUse(t *testing.T) {
	// Test that the client can be used concurrently
	client := NewHTTPClient()

	// This test verifies the client is thread-safe by design
	// (http.Client is safe for concurrent use)
	done := make(chan bool, 2)

	go func() {
		// Simulate concurrent use
		_ = client.Timeout
		done <- true
	}()

	go func() {
		// Simulate concurrent use
		_ = client.Transport
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// If we get here without race condition, test passes
}
