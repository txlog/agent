package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/viper"
	"github.com/txlog/agent/internal/client"
)

// setupTestServer creates a mock txlog server for testing
func setupTestServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/v1/machines":
			assets := []client.Asset{
				{MachineID: "abc123", Hostname: "server-01", OS: "AlmaLinux 9", AgentVersion: "1.0.0"},
				{MachineID: "def456", Hostname: "server-02", OS: "Rocky Linux 8", AgentVersion: "1.0.0"},
			}
			json.NewEncoder(w).Encode(assets)

		case "/v1/assets/requiring-restart":
			assets := []client.Asset{
				{MachineID: "abc123", Hostname: "server-01", NeedsRestarting: true},
			}
			json.NewEncoder(w).Encode(assets)

		case "/v1/transactions":
			transactions := []client.Transaction{
				{ID: 1, ExternalID: 100, MachineID: "abc123", Username: "root", Cmdline: "dnf update", ItemsCount: 5},
			}
			json.NewEncoder(w).Encode(transactions)

		case "/v1/items":
			items := []client.TransactionItem{
				{ID: 1, TransactionID: 100, Action: "Update", Package: "httpd", Version: "2.4.58", Release: "1.el9", Arch: "x86_64"},
			}
			json.NewEncoder(w).Encode(items)

		case "/v1/version":
			json.NewEncoder(w).Encode(map[string]string{"version": "1.18.0"})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestHandleListAssets(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	viper.Reset()
	viper.Set("server.url", server.URL)

	txlogClient := client.New()

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleListAssets(context.Background(), req, txlogClient)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error: %v", result)
	}

	// Check that result contains expected content
	content := result.Content[0].(mcp.TextContent)
	if content.Text == "" {
		t.Error("expected non-empty result text")
	}
}

func TestHandleListAssetsWithOSFilter(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	viper.Reset()
	viper.Set("server.url", server.URL)

	txlogClient := client.New()

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"os": "AlmaLinux",
	}

	result, err := handleListAssets(context.Background(), req, txlogClient)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error")
	}
}

func TestHandleGetRestartRequired(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	viper.Reset()
	viper.Set("server.url", server.URL)

	txlogClient := client.New()

	req := mcp.CallToolRequest{}

	result, err := handleGetRestartRequired(context.Background(), req, txlogClient)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error")
	}

	content := result.Content[0].(mcp.TextContent)
	if content.Text == "" {
		t.Error("expected non-empty result text")
	}
}

func TestHandleListTransactions(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	viper.Reset()
	viper.Set("server.url", server.URL)

	txlogClient := client.New()

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"machine_id": "abc123",
		"limit":      float64(10),
	}

	result, err := handleListTransactions(context.Background(), req, txlogClient)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error")
	}
}

func TestHandleGetTransactionDetails(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	viper.Reset()
	viper.Set("server.url", server.URL)

	txlogClient := client.New()

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"transaction_id": float64(100),
	}

	result, err := handleGetTransactionDetails(context.Background(), req, txlogClient)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("expected success, got error")
	}
}

func TestFormatAssets(t *testing.T) {
	assets := []client.Asset{
		{MachineID: "abc123", Hostname: "server-01", OS: "AlmaLinux 9", AgentVersion: "1.0.0", NeedsRestarting: false},
		{MachineID: "def456", Hostname: "server-02", OS: "AlmaLinux 9", AgentVersion: "1.0.0", NeedsRestarting: true},
	}

	result := formatAssets(assets)

	if result == "" {
		t.Error("expected non-empty formatted output")
	}

	// Check for expected content
	if !containsString(result, "Total de assets: 2") {
		t.Error("expected total count in output")
	}
	if !containsString(result, "server-01") {
		t.Error("expected hostname in output")
	}
}

func TestFormatTransactionItems(t *testing.T) {
	items := []client.TransactionItem{
		{Action: "Update", Package: "httpd", Version: "2.4.58", Release: "1.el9", Arch: "x86_64"},
		{Action: "Install", Package: "nginx", Version: "1.24.0", Release: "1.el9", Arch: "x86_64"},
	}

	result := formatTransactionItems(items)

	if result == "" {
		t.Error("expected non-empty formatted output")
	}
	if !containsString(result, "httpd") {
		t.Error("expected package name in output")
	}
}

func TestExtractMachineID(t *testing.T) {
	tests := []struct {
		uri      string
		expected string
	}{
		{"txlog://assets/abc123/transactions", "abc123"},
		{"txlog://assets/def456/executions", "def456"},
		{"txlog://assets/machine-id-123", "machine-id-123"},
		{"txlog://assets/", ""},
		{"txlog://other", ""},
	}

	for _, tt := range tests {
		result := extractMachineID(tt.uri)
		if result != tt.expected {
			t.Errorf("extractMachineID(%q) = %q, expected %q", tt.uri, result, tt.expected)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
