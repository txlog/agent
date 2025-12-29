package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
)

func TestNew(t *testing.T) {
	viper.Reset()
	viper.Set("server.url", "http://localhost:8080")

	client := New()

	if client == nil {
		t.Fatal("expected client to be created")
	}
	if client.baseURL != "http://localhost:8080" {
		t.Errorf("expected baseURL to be http://localhost:8080, got %s", client.baseURL)
	}
}

func TestListAssets(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/machines" {
			t.Errorf("expected path /v1/machines, got %s", r.URL.Path)
		}

		assets := []Asset{
			{
				MachineID:    "abc123",
				Hostname:     "server-01",
				OS:           "AlmaLinux 9",
				AgentVersion: "1.0.0",
			},
			{
				MachineID:    "def456",
				Hostname:     "server-02",
				OS:           "Rocky Linux 8",
				AgentVersion: "1.0.0",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(assets)
	}))
	defer server.Close()

	viper.Reset()
	viper.Set("server.url", server.URL)

	client := New()
	assets, err := client.ListAssets()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(assets) != 2 {
		t.Errorf("expected 2 assets, got %d", len(assets))
	}
	if assets[0].Hostname != "server-01" {
		t.Errorf("expected hostname server-01, got %s", assets[0].Hostname)
	}
}

func TestGetAssetsRequiringRestart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/assets/requiring-restart" {
			t.Errorf("expected path /v1/assets/requiring-restart, got %s", r.URL.Path)
		}

		assets := []Asset{
			{
				MachineID:       "abc123",
				Hostname:        "server-01",
				NeedsRestarting: true,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(assets)
	}))
	defer server.Close()

	viper.Reset()
	viper.Set("server.url", server.URL)

	client := New()
	assets, err := client.GetAssetsRequiringRestart()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(assets) != 1 {
		t.Errorf("expected 1 asset, got %d", len(assets))
	}
	if !assets[0].NeedsRestarting {
		t.Error("expected asset to need restarting")
	}
}

func TestGetTransactions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/transactions" {
			t.Errorf("expected path /v1/transactions, got %s", r.URL.Path)
		}

		machineID := r.URL.Query().Get("machine_id")
		if machineID != "abc123" {
			t.Errorf("expected machine_id abc123, got %s", machineID)
		}

		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("expected limit 10, got %s", limit)
		}

		transactions := []Transaction{
			{
				ID:         1,
				ExternalID: 100,
				MachineID:  "abc123",
				Username:   "root",
				Cmdline:    "dnf update",
				ItemsCount: 5,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transactions)
	}))
	defer server.Close()

	viper.Reset()
	viper.Set("server.url", server.URL)

	client := New()
	transactions, err := client.GetTransactions("abc123", 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(transactions))
	}
	if transactions[0].Cmdline != "dnf update" {
		t.Errorf("expected cmdline 'dnf update', got %s", transactions[0].Cmdline)
	}
}

func TestGetTransactionItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/items" {
			t.Errorf("expected path /v1/items, got %s", r.URL.Path)
		}

		items := []TransactionItem{
			{
				ID:            1,
				TransactionID: 100,
				Action:        "Update",
				Package:       "httpd",
				Version:       "2.4.58",
				Release:       "1.el9",
				Arch:          "x86_64",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
	}))
	defer server.Close()

	viper.Reset()
	viper.Set("server.url", server.URL)

	client := New()
	items, err := client.GetTransactionItems(100)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
	if items[0].Package != "httpd" {
		t.Errorf("expected package 'httpd', got %s", items[0].Package)
	}
}

func TestGetServerVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/version" {
			t.Errorf("expected path /v1/version, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"version": "1.18.0"})
	}))
	defer server.Close()

	viper.Reset()
	viper.Set("server.url", server.URL)

	client := New()
	version, err := client.GetServerVersion()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "1.18.0" {
		t.Errorf("expected version 1.18.0, got %s", version)
	}
}

func TestListAssetsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	viper.Reset()
	viper.Set("server.url", server.URL)

	client := New()
	_, err := client.ListAssets()

	if err == nil {
		t.Fatal("expected error for server error response")
	}
}

func TestGetExecutions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/executions" {
			t.Errorf("expected path /v1/executions, got %s", r.URL.Path)
		}

		executions := []Execution{
			{
				ID:              1,
				MachineID:       "abc123",
				Hostname:        "server-01",
				OS:              "AlmaLinux 9",
				AgentVersion:    "1.0.0",
				NeedsRestarting: false,
				ExecutedAt:      "2024-12-29T10:00:00Z",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(executions)
	}))
	defer server.Close()

	viper.Reset()
	viper.Set("server.url", server.URL)

	client := New()
	executions, err := client.GetExecutions("abc123", 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(executions) != 1 {
		t.Errorf("expected 1 execution, got %d", len(executions))
	}
}
