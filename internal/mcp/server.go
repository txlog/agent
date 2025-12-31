// Package mcp provides an MCP (Model Context Protocol) server implementation
// that exposes txlog server data to LLMs.
package mcp

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/txlog/agent/internal/client"
)

// NewServer creates a new MCP server configured with txlog tools, resources, and prompts.
// If compatibilityErr is not nil, it indicates the server version is incompatible,
// and all tools will return friendly error messages instead of executing.
func NewServer(compatibilityErr error) *server.MCPServer {
	s := server.NewMCPServer(
		"Txlog MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
	)

	// Create the txlog client
	txlogClient := client.New()

	// Register Tools
	registerTools(s, txlogClient, compatibilityErr)

	// Register Resources
	registerResources(s, txlogClient)

	// Register Prompts
	registerPrompts(s)

	return s
}
