package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/txlog/agent/internal/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP (Model Context Protocol) server commands",
	Long:  `Commands for running the txlog agent as an MCP server, enabling LLMs to query datacenter information.`,
}

var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server",
	Long: `Start the MCP server to expose txlog data to LLMs.

The server supports two transport modes:
  - stdio: Communicates via standard input/output (default, for Claude Desktop)
  - sse: HTTP Server-Sent Events for web-based clients

Examples:
  # Start MCP server in stdio mode (default)
  txlog mcp serve

  # Start MCP server with SSE transport on port 3000
  txlog mcp serve --transport sse --port 3000`,
	Run: func(cmd *cobra.Command, args []string) {
		runMCPServer()
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.AddCommand(mcpServeCmd)

	// Transport mode flag
	mcpServeCmd.Flags().String("transport", "stdio", "Transport mode: stdio or sse")
	viper.BindPFlag("mcp.transport", mcpServeCmd.Flags().Lookup("transport"))

	// SSE port flag
	mcpServeCmd.Flags().Int("port", 3000, "Port for SSE transport")
	viper.BindPFlag("mcp.port", mcpServeCmd.Flags().Lookup("port"))
}

func runMCPServer() {
	// Create the MCP server
	mcpServer := mcp.NewServer()

	transport := viper.GetString("mcp.transport")

	switch transport {
	case "stdio":
		// Create stdio transport with context, stdin, and stdout
		stdioTransport := server.NewStdioServer(mcpServer)
		if err := stdioTransport.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
	case "sse":
		port := viper.GetInt("mcp.port")
		sseTransport := server.NewSSEServer(mcpServer)
		fmt.Fprintf(os.Stderr, "Starting MCP SSE server on port %d...\n", port)
		if err := sseTransport.Start(fmt.Sprintf(":%d", port)); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown transport mode: %s\n", transport)
		os.Exit(1)
	}
}
