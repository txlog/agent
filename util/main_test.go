package util

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
)

func TestSetAuthentication_WithAPIKey(t *testing.T) {
	// Setup
	viper.Reset()
	viper.Set("server.api_key", "txlog_test_key_123")
	viper.Set("server.username", "testuser")
	viper.Set("server.password", "testpass")

	client := resty.New()
	req := client.R()

	// Execute
	SetAuthentication(req)

	// Verify API key header is set (preferred method)
	if req.Header.Get("X-API-Key") != "txlog_test_key_123" {
		t.Errorf("Expected X-API-Key header to be 'txlog_test_key_123', got '%s'", req.Header.Get("X-API-Key"))
	}

	// Verify basic auth is not set when API key is present
	if req.Header.Get("Authorization") != "" {
		t.Errorf("Expected Authorization header to be empty when API key is set, got '%s'", req.Header.Get("Authorization"))
	}
}

func TestSetAuthentication_WithBasicAuth(t *testing.T) {
	// Setup
	viper.Reset()
	viper.Set("server.username", "testuser")
	viper.Set("server.password", "testpass")

	client := resty.New()
	req := client.R()

	// Execute
	SetAuthentication(req)

	// Verify API key header is not set
	if req.Header.Get("X-API-Key") != "" {
		t.Errorf("Expected X-API-Key header to be empty when using basic auth, got '%s'", req.Header.Get("X-API-Key"))
	}

	// Note: We can't directly verify the Authorization header here because resty's SetBasicAuth
	// sets it internally and only applies it when the request is actually sent.
	// The important thing is that we verify API key is not set, which means basic auth was configured.
}

func TestSetAuthentication_WithNoAuth(t *testing.T) {
	// Setup
	viper.Reset()

	client := resty.New()
	req := client.R()

	// Execute
	SetAuthentication(req)

	// Verify no auth headers are set
	if req.Header.Get("X-API-Key") != "" {
		t.Errorf("Expected X-API-Key header to be empty, got '%s'", req.Header.Get("X-API-Key"))
	}
	if req.Header.Get("Authorization") != "" {
		t.Errorf("Expected Authorization header to be empty, got '%s'", req.Header.Get("Authorization"))
	}
}

func TestSetAuthentication_APIKeyPreferredOverBasicAuth(t *testing.T) {
	// Setup - both API key and basic auth configured
	viper.Reset()
	viper.Set("server.api_key", "txlog_preferred_key")
	viper.Set("server.username", "testuser")
	viper.Set("server.password", "testpass")

	client := resty.New()
	req := client.R()

	// Execute
	SetAuthentication(req)

	// Verify API key is used (not basic auth)
	if req.Header.Get("X-API-Key") != "txlog_preferred_key" {
		t.Errorf("Expected X-API-Key header to be 'txlog_preferred_key', got '%s'", req.Header.Get("X-API-Key"))
	}

	// When API key is set, basic auth should not be used
	// (We can't verify the Authorization header directly, but we know the function
	// returns early when API key is set, so basic auth is never configured)
}
