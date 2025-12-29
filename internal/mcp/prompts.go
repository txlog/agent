package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerPrompts registers all MCP prompts with the server.
func registerPrompts(s *server.MCPServer) {
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
}
