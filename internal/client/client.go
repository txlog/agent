// Package client provides an HTTP client for interacting with the txlog server API.
package client

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
	"github.com/txlog/agent/util"
)

// Asset represents a server/machine registered in the txlog system.
type Asset struct {
	MachineID       string `json:"machine_id"`
	Hostname        string `json:"hostname"`
	OS              string `json:"os"`
	AgentVersion    string `json:"agent_version"`
	NeedsRestarting bool   `json:"needs_restarting"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// Transaction represents a package transaction on an asset.
type Transaction struct {
	ID         int    `json:"id"`
	ExternalID int    `json:"external_id"`
	MachineID  string `json:"machine_id"`
	Username   string `json:"username"`
	Cmdline    string `json:"cmdline"`
	ExecutedAt string `json:"executed_at"`
	ItemsCount int    `json:"items_count"`
}

// TransactionItem represents an individual package change within a transaction.
type TransactionItem struct {
	ID            int    `json:"id"`
	TransactionID int    `json:"transaction_id"`
	Action        string `json:"action"`
	Package       string `json:"package"`
	Version       string `json:"version"`
	Release       string `json:"release"`
	Epoch         string `json:"epoch"`
	Arch          string `json:"arch"`
}

// Execution represents an agent execution record.
type Execution struct {
	ID              int    `json:"id"`
	MachineID       string `json:"machine_id"`
	Hostname        string `json:"hostname"`
	OS              string `json:"os"`
	AgentVersion    string `json:"agent_version"`
	NeedsRestarting bool   `json:"needs_restarting"`
	ExecutedAt      string `json:"executed_at"`
}

// PackageInfo represents package information across assets.
type PackageInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Release string `json:"release"`
	Arch    string `json:"arch"`
	Count   int    `json:"count"`
}

// Client is the HTTP client for the txlog server API.
type Client struct {
	baseURL    string
	httpClient *resty.Client
}

// New creates a new txlog server API client.
func New() *Client {
	baseURL := viper.GetString("server.url")
	httpClient := resty.New()
	httpClient.SetBaseURL(baseURL)

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// newRequest creates a new request with authentication configured.
func (c *Client) newRequest() *resty.Request {
	req := c.httpClient.R()
	util.SetAuthentication(req)
	return req
}

// ListAssets retrieves all assets from the server.
func (c *Client) ListAssets() ([]Asset, error) {
	resp, err := c.newRequest().Get("/v1/machines")
	if err != nil {
		return nil, fmt.Errorf("failed to list assets: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var assets []Asset
	if err := json.Unmarshal(resp.Body(), &assets); err != nil {
		return nil, fmt.Errorf("failed to parse assets response: %w", err)
	}

	return assets, nil
}

// GetAssetsRequiringRestart retrieves assets that need to be restarted.
func (c *Client) GetAssetsRequiringRestart() ([]Asset, error) {
	resp, err := c.newRequest().Get("/v1/assets/requiring-restart")
	if err != nil {
		return nil, fmt.Errorf("failed to get assets requiring restart: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var assets []Asset
	if err := json.Unmarshal(resp.Body(), &assets); err != nil {
		return nil, fmt.Errorf("failed to parse assets response: %w", err)
	}

	return assets, nil
}

// GetTransactions retrieves transactions for a specific machine.
func (c *Client) GetTransactions(machineID string, limit int) ([]Transaction, error) {
	req := c.newRequest()
	req.SetQueryParam("machine_id", machineID)
	if limit > 0 {
		req.SetQueryParam("limit", fmt.Sprintf("%d", limit))
	}

	resp, err := req.Get("/v1/transactions")
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var transactions []Transaction
	if err := json.Unmarshal(resp.Body(), &transactions); err != nil {
		return nil, fmt.Errorf("failed to parse transactions response: %w", err)
	}

	return transactions, nil
}

// GetTransactionItems retrieves items for a specific transaction.
func (c *Client) GetTransactionItems(transactionID int) ([]TransactionItem, error) {
	resp, err := c.newRequest().Get(fmt.Sprintf("/v1/items?transaction_id=%d", transactionID))
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction items: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var items []TransactionItem
	if err := json.Unmarshal(resp.Body(), &items); err != nil {
		return nil, fmt.Errorf("failed to parse items response: %w", err)
	}

	return items, nil
}

// GetExecutions retrieves execution history for a machine.
func (c *Client) GetExecutions(machineID string, limit int) ([]Execution, error) {
	req := c.newRequest()
	req.SetQueryParam("machine_id", machineID)
	if limit > 0 {
		req.SetQueryParam("limit", fmt.Sprintf("%d", limit))
	}

	resp, err := req.Get("/v1/executions")
	if err != nil {
		return nil, fmt.Errorf("failed to get executions: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var executions []Execution
	if err := json.Unmarshal(resp.Body(), &executions); err != nil {
		return nil, fmt.Errorf("failed to parse executions response: %w", err)
	}

	return executions, nil
}

// SearchPackageAssets finds assets using a specific package.
func (c *Client) SearchPackageAssets(name, version, release string) ([]Asset, error) {
	url := fmt.Sprintf("/v1/packages/%s/%s/%s/assets", name, version, release)
	resp, err := c.newRequest().Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to search package assets: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var assets []Asset
	if err := json.Unmarshal(resp.Body(), &assets); err != nil {
		return nil, fmt.Errorf("failed to parse assets response: %w", err)
	}

	return assets, nil
}

// GetServerVersion retrieves the server version.
func (c *Client) GetServerVersion() (string, error) {
	resp, err := c.newRequest().Get("/v1/version")
	if err != nil {
		return "", fmt.Errorf("failed to get server version: %w", err)
	}

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var result struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "", fmt.Errorf("failed to parse version response: %w", err)
	}

	return result.Version, nil
}

// MonthlyReportPackage represents a package update in the monthly report.
type MonthlyReportPackage struct {
	OSVersion      string `json:"os_version"`
	PackageRPM     string `json:"package_rpm"`
	AssetsAffected int    `json:"assets_affected"`
}

// MonthlyReportResponse represents the response from the monthly report endpoint.
type MonthlyReportResponse struct {
	AssetCount int                    `json:"asset_count"`
	Month      int                    `json:"month"`
	Year       int                    `json:"year"`
	Packages   []MonthlyReportPackage `json:"packages"`
}

// GetMonthlyReport retrieves the monthly package update report for a specific month/year.
func (c *Client) GetMonthlyReport(month, year int) (*MonthlyReportResponse, error) {
	resp, err := c.newRequest().
		SetQueryParam("month", fmt.Sprintf("%d", month)).
		SetQueryParam("year", fmt.Sprintf("%d", year)).
		Get("/v1/reports/monthly")
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly report: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var report MonthlyReportResponse
	if err := json.Unmarshal(resp.Body(), &report); err != nil {
		return nil, fmt.Errorf("failed to parse monthly report response: %w", err)
	}

	return &report, nil
}

// GetAssetByHostname finds an asset by its hostname.
func (c *Client) GetAssetByHostname(hostname string) (*Asset, error) {
	resp, err := c.newRequest().
		SetQueryParam("hostname", hostname).
		Get("/v1/machines/ids")
	if err != nil {
		return nil, fmt.Errorf("failed to get asset by hostname: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var machineIDs []string
	if err := json.Unmarshal(resp.Body(), &machineIDs); err != nil {
		return nil, fmt.Errorf("failed to parse machine IDs response: %w", err)
	}

	if len(machineIDs) == 0 {
		return nil, fmt.Errorf("asset not found: %s", hostname)
	}

	// Get the full asset details for the first machine ID
	assets, err := c.ListAssets()
	if err != nil {
		return nil, err
	}

	for _, asset := range assets {
		if asset.MachineID == machineIDs[0] {
			return &asset, nil
		}
	}

	return nil, fmt.Errorf("asset not found: %s", hostname)
}
