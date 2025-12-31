package mcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/txlog/agent/internal/client"
)

// registerPrompts registers all MCP prompts with the server.
func registerPrompts(s *server.MCPServer, txlogClient *client.Client) {
	// Prompt: infrastructure_report
	s.AddPrompt(
		mcp.NewPrompt("infrastructure_report",
			mcp.WithPromptDescription("Generate a complete infrastructure report using the available tools."),
		),
		func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return &mcp.GetPromptResult{
				Description: "Infrastructure Report",
				Messages: []mcp.PromptMessage{
					{
						Role: mcp.RoleUser,
						Content: mcp.TextContent{
							Type: "text",
							Text: `Please generate a complete infrastructure report using the available tools.

The report should include:

1. **Overview**
   - Total of servers
   - Distribution by OS
   - Agent versions in use

2. **Health Status**
   - Servers that need to restart
   - Latest package updates

3. **Recommendations**
   - Suggested maintenance actions
   - Security alerts if any

Use the tools list_assets and get_restart_required to obtain the necessary data.`,
						},
					},
				},
			}, nil
		},
	)

	// Prompt: security_audit
	s.AddPrompt(
		mcp.NewPrompt("security_audit",
			mcp.WithPromptDescription("Perform a security audit focused on packages."),
			mcp.WithArgument("package",
				mcp.ArgumentDescription("Package name to audit (optional)"),
			),
		),
		func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			// Arguments is map[string]string, access directly
			packageName := req.Params.Arguments["package"]

			promptText := `Please perform a security audit of the infrastructure.

Verify:

1. **Critical Security Packages**
   - openssl
   - openssh
   - kernel
   - glibc

2. **Version Consistency**
   - Identify servers with different versions of the same package
   - Highlight potentially vulnerable versions

3. **Pending Restart Servers**
   - List servers that need to restart after updates`

			if packageName != "" {
				promptText += "\n\n**Focus on package:** " + packageName
			}

			return &mcp.GetPromptResult{
				Description: "Security Audit",
				Messages: []mcp.PromptMessage{
					{
						Role: mcp.RoleUser,
						Content: mcp.TextContent{
							Type: "text",
							Text: promptText,
						},
					},
				},
			}, nil
		},
	)

	// Prompt: troubleshoot_asset
	s.AddPrompt(
		mcp.NewPrompt("troubleshoot_asset",
			mcp.WithPromptDescription("Troubleshooting guide for a specific server"),
			mcp.WithArgument("hostname",
				mcp.ArgumentDescription("Hostname of the server to analyze"),
				mcp.RequiredArgument(),
			),
		),
		func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			// Arguments is map[string]string, access directly
			hostname := req.Params.Arguments["hostname"]

			return &mcp.GetPromptResult{
				Description: "Troubleshooting guide for a specific server",
				Messages: []mcp.PromptMessage{
					{
						Role: mcp.RoleUser,
						Content: mcp.TextContent{
							Type: "text",
							Text: `Please analyze server "` + hostname + `" for troubleshooting.

Steps:

1. **Basic Information**
   - Use get_asset_details to get server information
   - Check if a restart is required

2. **Change History**
   - Use list_transactions to see the latest transactions
   - Identify recent changes that may have caused issues

3. **Analysis**
   - List packages that were recently changed
   - Use get_transaction_details to see details of suspicious transactions

4. **Recommendations**
   - Suggest corrective actions based on the analysis`,
						},
					},
				},
			}, nil
		},
	)

	// Prompt: compliance_check
	s.AddPrompt(
		mcp.NewPrompt("compliance_check",
			mcp.WithPromptDescription("Verification of compliance of the infrastructure"),
		),
		func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return &mcp.GetPromptResult{
				Description: "Verification of Compliance",
				Messages: []mcp.PromptMessage{
					{
						Role: mcp.RoleUser,
						Content: mcp.TextContent{
							Type: "text",
							Text: `Please perform a compliance check of the infrastructure.

Verify:

1. **Version Standardization**
   - All production servers should have the same version of critical packages
   - Identify version deviations

2. **Pending Updates**
   - Servers that need to restart
   - Time since the last update

3. **Documentation**
   - List servers by OS
   - Identify legacy operating systems

Generate a report in table format when possible.`,
						},
					},
				},
			}, nil
		},
	)

	// Prompt: executive_report
	s.AddPrompt(
		mcp.NewPrompt("executive_report",
			mcp.WithPromptDescription("Generate a monthly executive report for management about package updates"),
			mcp.WithArgument("month",
				mcp.ArgumentDescription("Month (1-12) for the report"),
				mcp.RequiredArgument(),
			),
			mcp.WithArgument("year",
				mcp.ArgumentDescription("Year (e.g., 2024) for the report"),
				mcp.RequiredArgument(),
			),
		),
		func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			monthStr := req.Params.Arguments["month"]
			yearStr := req.Params.Arguments["year"]

			if monthStr == "" || yearStr == "" {
				return &mcp.GetPromptResult{
					Description: "Executive Report Error",
					Messages: []mcp.PromptMessage{
						{
							Role: mcp.RoleUser,
							Content: mcp.TextContent{
								Type: "text",
								Text: "Error: Both 'month' and 'year' parameters are required for the executive report.",
							},
						},
					},
				}, nil
			}

			month, err := strconv.Atoi(monthStr)
			if err != nil || month < 1 || month > 12 {
				return &mcp.GetPromptResult{
					Description: "Executive Report Error",
					Messages: []mcp.PromptMessage{
						{
							Role: mcp.RoleUser,
							Content: mcp.TextContent{
								Type: "text",
								Text: "Error: Month must be a valid number between 1 and 12.",
							},
						},
					},
				}, nil
			}

			year, err := strconv.Atoi(yearStr)
			if err != nil || year < 2000 || year > 2100 {
				return &mcp.GetPromptResult{
					Description: "Executive Report Error",
					Messages: []mcp.PromptMessage{
						{
							Role: mcp.RoleUser,
							Content: mcp.TextContent{
								Type: "text",
								Text: "Error: Year must be a valid number between 2000 and 2100.",
							},
						},
					},
				}, nil
			}

			// Fetch data from the server
			report, err := txlogClient.GetMonthlyReport(month, year)
			if err != nil {
				return &mcp.GetPromptResult{
					Description: "Executive Report Error",
					Messages: []mcp.PromptMessage{
						{
							Role: mcp.RoleUser,
							Content: mcp.TextContent{
								Type: "text",
								Text: fmt.Sprintf("Error fetching report data from server: %v", err),
							},
						},
					},
				}, nil
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

			return &mcp.GetPromptResult{
				Description: fmt.Sprintf("Monthly Executive Report - %s %d", monthName, year),
				Messages: []mcp.PromptMessage{
					{
						Role: mcp.RoleUser,
						Content: mcp.TextContent{
							Type: "text",
							Text: promptText,
						},
					},
				},
			}, nil
		},
	)
}
