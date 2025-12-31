package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/txlog/agent/internal/client"
)

// registerTools registers all MCP tools with the server.
// If compatibilityErr is not nil, all tools will return an error message
// indicating the server version is incompatible.
func registerTools(s *server.MCPServer, txlogClient *client.Client, compatibilityErr error) {
	// Helper function to wrap handlers with compatibility check
	wrapHandler := func(handler func(context.Context, mcp.CallToolRequest, *client.Client) (*mcp.CallToolResult, error)) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if compatibilityErr != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Server compatibility error: %v", compatibilityErr)), nil
			}
			return handler(ctx, req, txlogClient)
		}
	}

	// Tool: list_assets
	listAssetsTool := mcp.NewTool("list_assets",
		mcp.WithDescription("Lists all assets (servers) in the datacenter. Use to get an overview of the infrastructure."),
		mcp.WithString("os",
			mcp.Description("Filter by operating system (e.g., 'AlmaLinux 9', 'Rocky Linux 8')"),
		),
		mcp.WithString("agent_version",
			mcp.Description("Filter by txlog agent version"),
		),
	)

	s.AddTool(listAssetsTool, wrapHandler(handleListAssets))

	// Tool: get_asset_details
	getAssetDetailsTool := mcp.NewTool("get_asset_details",
		mcp.WithDescription("Gets details of a specific asset by hostname or machine_id"),
		mcp.WithString("hostname",
			mcp.Description("Server hostname"),
		),
		mcp.WithString("machine_id",
			mcp.Description("Server machine ID"),
		),
	)

	s.AddTool(getAssetDetailsTool, wrapHandler(handleGetAssetDetails))

	// Tool: list_transactions
	listTransactionsTool := mcp.NewTool("list_transactions",
		mcp.WithDescription("Lists package transactions (dnf/yum) for an asset. Useful for troubleshooting and auditing."),
		mcp.WithString("machine_id",
			mcp.Description("Server machine ID (required)"),
			mcp.Required(),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of transactions to return (default: 10)"),
		),
	)

	s.AddTool(listTransactionsTool, wrapHandler(handleListTransactions))

	// Tool: get_restart_required
	restartTool := mcp.NewTool("get_restart_required",
		mcp.WithDescription("Lists all assets that need to be restarted after package updates"),
	)

	s.AddTool(restartTool, wrapHandler(handleGetRestartRequired))

	// Tool: search_package
	searchPackageTool := mcp.NewTool("search_package",
		mcp.WithDescription("Searches which assets have a specific package installed"),
		mcp.WithString("name",
			mcp.Description("Package name (e.g., 'openssl', 'httpd')"),
			mcp.Required(),
		),
		mcp.WithString("version",
			mcp.Description("Specific package version"),
		),
		mcp.WithString("release",
			mcp.Description("Specific package release"),
		),
	)

	s.AddTool(searchPackageTool, wrapHandler(handleSearchPackage))

	// Tool: get_transaction_details
	getTransactionDetailsTool := mcp.NewTool("get_transaction_details",
		mcp.WithDescription("Gets the details of a specific transaction, including all changed packages"),
		mcp.WithNumber("transaction_id",
			mcp.Description("Transaction ID"),
			mcp.Required(),
		),
	)

	s.AddTool(getTransactionDetailsTool, wrapHandler(handleGetTransactionDetails))

	// Tool: generate_executive_report
	generateExecutiveReportTool := mcp.NewTool("generate_executive_report",
		mcp.WithDescription("Generates a monthly executive report for management about package updates. Returns data and instructions for creating a professional report highlighting security updates, CVEs, and infrastructure impact."),
		mcp.WithNumber("month",
			mcp.Description("Month (1-12) for the report"),
			mcp.Required(),
		),
		mcp.WithNumber("year",
			mcp.Description("Year (e.g., 2024) for the report"),
			mcp.Required(),
		),
	)

	s.AddTool(generateExecutiveReportTool, wrapHandler(handleGenerateExecutiveReport))
}

// handleListAssets handles the list_assets tool call.
func handleListAssets(_ context.Context, req mcp.CallToolRequest, txlogClient *client.Client) (*mcp.CallToolResult, error) {
	assets, err := txlogClient.ListAssets()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error listing assets: %v", err)), nil
	}

	// Filter by OS if specified
	osFilter := req.GetString("os", "")
	agentVersionFilter := req.GetString("agent_version", "")

	var filtered []client.Asset
	for _, asset := range assets {
		if osFilter != "" && !strings.Contains(strings.ToLower(asset.OS), strings.ToLower(osFilter)) {
			continue
		}
		if agentVersionFilter != "" && asset.AgentVersion != agentVersionFilter {
			continue
		}
		filtered = append(filtered, asset)
	}

	if len(filtered) == 0 {
		return mcp.NewToolResultText("No assets found with the specified filters."), nil
	}

	return mcp.NewToolResultText(formatAssets(filtered)), nil
}

// handleGetAssetDetails handles the get_asset_details tool call.
func handleGetAssetDetails(_ context.Context, req mcp.CallToolRequest, txlogClient *client.Client) (*mcp.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	machineID := req.GetString("machine_id", "")

	if hostname == "" && machineID == "" {
		return mcp.NewToolResultError("Either hostname or machine_id must be provided"), nil
	}

	var asset *client.Asset
	var err error

	if hostname != "" {
		asset, err = txlogClient.GetAssetByHostname(hostname)
	} else {
		// Search by machine_id
		assets, listErr := txlogClient.ListAssets()
		if listErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error fetching asset: %v", listErr)), nil
		}
		for _, a := range assets {
			if a.MachineID == machineID {
				asset = &a
				break
			}
		}
		if asset == nil {
			err = fmt.Errorf("asset not found: %s", machineID)
		}
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
	}

	// Get recent transactions for context
	transactions, _ := txlogClient.GetTransactions(asset.MachineID, 5)

	return mcp.NewToolResultText(formatAssetDetails(asset, transactions)), nil
}

// handleListTransactions handles the list_transactions tool call.
func handleListTransactions(_ context.Context, req mcp.CallToolRequest, txlogClient *client.Client) (*mcp.CallToolResult, error) {
	machineID := req.GetString("machine_id", "")
	limit := req.GetInt("limit", 0)
	if limit <= 0 {
		limit = 10
	}

	transactions, err := txlogClient.GetTransactions(machineID, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error listing transactions: %v", err)), nil
	}

	if len(transactions) == 0 {
		return mcp.NewToolResultText("No transactions found for this asset."), nil
	}

	return mcp.NewToolResultText(formatTransactions(transactions)), nil
}

// handleGetRestartRequired handles the get_restart_required tool call.
func handleGetRestartRequired(_ context.Context, req mcp.CallToolRequest, txlogClient *client.Client) (*mcp.CallToolResult, error) {
	assets, err := txlogClient.GetAssetsRequiringRestart()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error fetching assets: %v", err)), nil
	}

	if len(assets) == 0 {
		return mcp.NewToolResultText("✅ No assets need to be restarted at this time."), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("⚠️ **%d assets need to be restarted:**\n\n", len(assets)))
	for _, asset := range assets {
		sb.WriteString(fmt.Sprintf("- **%s** (%s)\n", asset.Hostname, asset.OS))
	}
	sb.WriteString("\nRecommendation: Schedule a maintenance window to restart these servers.")

	return mcp.NewToolResultText(sb.String()), nil
}

// handleSearchPackage handles the search_package tool call.
func handleSearchPackage(_ context.Context, req mcp.CallToolRequest, txlogClient *client.Client) (*mcp.CallToolResult, error) {
	name := req.GetString("name", "")
	version := req.GetString("version", "")
	release := req.GetString("release", "")

	if version == "" {
		version = "*"
	}
	if release == "" {
		release = "*"
	}

	assets, err := txlogClient.SearchPackageAssets(name, version, release)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error searching package: %v", err)), nil
	}

	if len(assets) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No assets found with package %s installed.", name)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**Package %s found on %d assets:**\n\n", name, len(assets)))
	for _, asset := range assets {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", asset.Hostname, asset.OS))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// handleGetTransactionDetails handles the get_transaction_details tool call.
func handleGetTransactionDetails(_ context.Context, req mcp.CallToolRequest, txlogClient *client.Client) (*mcp.CallToolResult, error) {
	transactionID := req.GetInt("transaction_id", 0)

	items, err := txlogClient.GetTransactionItems(transactionID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error fetching transaction details: %v", err)), nil
	}

	if len(items) == 0 {
		return mcp.NewToolResultText("No items found for this transaction."), nil
	}

	return mcp.NewToolResultText(formatTransactionItems(items)), nil
}

// formatAssets formats a list of assets for display.
func formatAssets(assets []client.Asset) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**Total assets: %d**\n\n", len(assets)))

	// Group by OS
	osCounts := make(map[string]int)
	for _, asset := range assets {
		osCounts[asset.OS]++
	}

	sb.WriteString("**Distribution by OS:**\n")
	for os, count := range osCounts {
		sb.WriteString(fmt.Sprintf("- %s: %d\n", os, count))
	}

	sb.WriteString("\n**Asset list:**\n")
	for _, asset := range assets {
		status := "✅"
		if asset.NeedsRestarting {
			status = "⚠️ (needs restart)"
		}
		sb.WriteString(fmt.Sprintf("- **%s** | %s | Agent v%s %s\n",
			asset.Hostname, asset.OS, asset.AgentVersion, status))
	}

	return sb.String()
}

// formatAssetDetails formats asset details for display.
func formatAssetDetails(asset *client.Asset, transactions []client.Transaction) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Asset: %s\n\n", asset.Hostname))
	sb.WriteString(fmt.Sprintf("- **Machine ID:** `%s`\n", asset.MachineID))
	sb.WriteString(fmt.Sprintf("- **Operating System:** %s\n", asset.OS))
	sb.WriteString(fmt.Sprintf("- **Agent Version:** %s\n", asset.AgentVersion))

	if asset.NeedsRestarting {
		sb.WriteString("- **Status:** ⚠️ Needs restart\n")
	} else {
		sb.WriteString("- **Status:** ✅ OK\n")
	}

	if len(transactions) > 0 {
		sb.WriteString("\n### Recent Transactions:\n")
		for _, t := range transactions {
			sb.WriteString(fmt.Sprintf("- [%s] %s - %d packages changed\n",
				t.ExecutedAt, t.Cmdline, t.ItemsCount))
		}
	}

	return sb.String()
}

// formatTransactions formats transactions for display.
func formatTransactions(transactions []client.Transaction) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**Last %d transactions:**\n\n", len(transactions)))

	for _, t := range transactions {
		sb.WriteString(fmt.Sprintf("### Transaction #%d\n", t.ExternalID))
		sb.WriteString(fmt.Sprintf("- **Date:** %s\n", t.ExecutedAt))
		sb.WriteString(fmt.Sprintf("- **User:** %s\n", t.Username))
		sb.WriteString(fmt.Sprintf("- **Command:** `%s`\n", t.Cmdline))
		sb.WriteString(fmt.Sprintf("- **Packages changed:** %d\n\n", t.ItemsCount))
	}

	return sb.String()
}

// formatTransactionItems formats transaction items for display.
func formatTransactionItems(items []client.TransactionItem) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**Transaction details (%d packages):**\n\n", len(items)))

	// Group by action
	actions := make(map[string][]client.TransactionItem)
	for _, item := range items {
		actions[item.Action] = append(actions[item.Action], item)
	}

	actionLabels := map[string]string{
		"Install":   "📦 Installed",
		"Update":    "🔄 Updated",
		"Remove":    "🗑️ Removed",
		"Downgrade": "⬇️ Downgraded",
		"Reinstall": "♻️ Reinstalled",
	}

	for action, actionItems := range actions {
		label := actionLabels[action]
		if label == "" {
			label = action
		}
		sb.WriteString(fmt.Sprintf("### %s (%d)\n", label, len(actionItems)))
		for _, item := range actionItems {
			version := item.Version
			if item.Epoch != "" {
				version = item.Epoch + ":" + version
			}
			sb.WriteString(fmt.Sprintf("- %s-%s-%s.%s\n", item.Package, version, item.Release, item.Arch))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// toJSON converts an object to JSON string for debugging.
func toJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

// handleGenerateExecutiveReport handles the generate_executive_report tool call.
func handleGenerateExecutiveReport(_ context.Context, req mcp.CallToolRequest, txlogClient *client.Client) (*mcp.CallToolResult, error) {
	month := req.GetInt("month", 0)
	year := req.GetInt("year", 0)

	if month < 1 || month > 12 {
		return mcp.NewToolResultError("Month must be a valid number between 1 and 12."), nil
	}

	if year < 2000 || year > 2100 {
		return mcp.NewToolResultError("Year must be a valid number between 2000 and 2100."), nil
	}

	// Fetch data from the server
	report, err := txlogClient.GetMonthlyReport(month, year)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error fetching report data from server: %v", err)), nil
	}

	// Get month name
	monthName := time.Month(month).String()

	// Build CSV content from packages
	var csvBuilder strings.Builder
	csvBuilder.WriteString("os_version,package_rpm,assets_affected\n")
	for _, pkg := range report.Packages {
		csvBuilder.WriteString(fmt.Sprintf("%s,%s,%d\n", pkg.OSVersion, pkg.PackageRPM, pkg.AssetsAffected))
	}
	csvContent := csvBuilder.String()

	promptText := fmt.Sprintf(`Act as an SRE specialist preparing an executive summary for management. The tone should be professional, direct, and focused on impact and security. Format the final response in Markdown.

Report Period: %s %d

Context: Below is the list of packages that were updated in our infrastructure during the reporting period. The data shows the number of servers where each package was updated, and the total number of update transactions (some servers may receive the same package update multiple times). Our total infrastructure consists of %d servers.

Data:
---
%s
---

Task:
Based on this data, write a brief management report in Markdown format (1-2 paragraphs) highlighting:
1. The most critical and high-impact updates, considering the number of affected servers. Give special attention to security packages (such as OpenSSL) or system packages (such as the Kernel).
2. The overall reach of the updates (percentage of servers impacted by the most important updates).
3. Any relevant patterns or observations that management should be aware of.
4. Research on the internet which CVEs these packages may have fixed during the selected period, always summarizing each CVE in one or two sentences. Use Red Hat Enterprise Linux errata as a reference, since RPM-based systems are based on RHEL.`, monthName, year, report.AssetCount, csvContent)

	return mcp.NewToolResultText(promptText), nil
}
