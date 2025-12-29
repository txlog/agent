package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/txlog/agent/internal/client"
)

// registerResources registers all MCP resources with the server.
func registerResources(s *server.MCPServer, txlogClient *client.Client) {
	// Resource: txlog://assets
	s.AddResource(
		mcp.NewResource(
			"txlog://assets",
			"Lista de todos os assets (servidores) do datacenter",
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			assets, err := txlogClient.ListAssets()
			if err != nil {
				return nil, fmt.Errorf("failed to list assets: %w", err)
			}

			data, err := json.MarshalIndent(assets, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal assets: %w", err)
			}

			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:  req.Params.URI,
					Text: string(data),
				},
			}, nil
		},
	)

	// Resource: txlog://assets/requiring-restart
	s.AddResource(
		mcp.NewResource(
			"txlog://assets/requiring-restart",
			"List of assets requiring restart",
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			assets, err := txlogClient.GetAssetsRequiringRestart()
			if err != nil {
				return nil, fmt.Errorf("failed to get assets requiring restart: %w", err)
			}

			data, err := json.MarshalIndent(assets, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal assets: %w", err)
			}

			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:  req.Params.URI,
					Text: string(data),
				},
			}, nil
		},
	)

	// Resource: txlog://version
	s.AddResource(
		mcp.NewResource(
			"txlog://version",
			"Txlog server version",
			mcp.WithMIMEType("text/plain"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			version, err := txlogClient.GetServerVersion()
			if err != nil {
				return nil, fmt.Errorf("failed to get server version: %w", err)
			}

			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:  req.Params.URI,
					Text: version,
				},
			}, nil
		},
	)

	// Resource Template: txlog://assets/{machine_id}/transactions
	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"txlog://assets/{machine_id}/transactions",
			"List of transactions for a specific asset",
			mcp.WithTemplateMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			// Extract machine_id from URI
			machineID := extractMachineID(req.Params.URI)
			if machineID == "" {
				return nil, fmt.Errorf("machine_id not found in URI")
			}

			transactions, err := txlogClient.GetTransactions(machineID, 50)
			if err != nil {
				return nil, fmt.Errorf("failed to get transactions: %w", err)
			}

			data, err := json.MarshalIndent(transactions, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal transactions: %w", err)
			}

			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:  req.Params.URI,
					Text: string(data),
				},
			}, nil
		},
	)

	// Resource Template: txlog://assets/{machine_id}/executions
	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"txlog://assets/{machine_id}/executions",
			"History of agent executions in an asset",
			mcp.WithTemplateMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			machineID := extractMachineID(req.Params.URI)
			if machineID == "" {
				return nil, fmt.Errorf("machine_id not found in URI")
			}

			executions, err := txlogClient.GetExecutions(machineID, 50)
			if err != nil {
				return nil, fmt.Errorf("failed to get executions: %w", err)
			}

			data, err := json.MarshalIndent(executions, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal executions: %w", err)
			}

			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:  req.Params.URI,
					Text: string(data),
				},
			}, nil
		},
	)
}

// extractMachineID extracts the machine_id from a URI like txlog://assets/{machine_id}/transactions
func extractMachineID(uri string) string {
	// URI format: txlog://assets/{machine_id}/...
	const prefix = "txlog://assets/"
	if len(uri) <= len(prefix) {
		return ""
	}

	rest := uri[len(prefix):]
	// Find the next slash or end of string
	for i, c := range rest {
		if c == '/' {
			return rest[:i]
		}
	}
	return rest
}
