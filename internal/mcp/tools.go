package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/txlog/agent/internal/client"
)

// registerTools registers all MCP tools with the server.
func registerTools(s *server.MCPServer, txlogClient *client.Client) {
	// Tool: list_assets
	listAssetsTool := mcp.NewTool("list_assets",
		mcp.WithDescription("Lista todos os assets (servidores) do datacenter. Use para obter uma visão geral da infraestrutura."),
		mcp.WithString("os",
			mcp.Description("Filtrar por sistema operacional (ex: 'AlmaLinux 9', 'Rocky Linux 8')"),
		),
		mcp.WithString("agent_version",
			mcp.Description("Filtrar por versão do agent txlog"),
		),
	)

	s.AddTool(listAssetsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleListAssets(ctx, req, txlogClient)
	})

	// Tool: get_asset_details
	getAssetDetailsTool := mcp.NewTool("get_asset_details",
		mcp.WithDescription("Obtém detalhes de um asset específico pelo hostname ou machine_id"),
		mcp.WithString("hostname",
			mcp.Description("Hostname do servidor"),
		),
		mcp.WithString("machine_id",
			mcp.Description("Machine ID do servidor"),
		),
	)

	s.AddTool(getAssetDetailsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleGetAssetDetails(ctx, req, txlogClient)
	})

	// Tool: list_transactions
	listTransactionsTool := mcp.NewTool("list_transactions",
		mcp.WithDescription("Lista transações de pacotes (dnf/yum) de um asset. Útil para troubleshooting e auditoria."),
		mcp.WithString("machine_id",
			mcp.Description("Machine ID do servidor (obrigatório)"),
			mcp.Required(),
		),
		mcp.WithNumber("limit",
			mcp.Description("Número máximo de transações a retornar (padrão: 10)"),
		),
	)

	s.AddTool(listTransactionsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleListTransactions(ctx, req, txlogClient)
	})

	// Tool: get_restart_required
	restartTool := mcp.NewTool("get_restart_required",
		mcp.WithDescription("Lista todos os assets que precisam ser reiniciados após atualizações de pacotes"),
	)

	s.AddTool(restartTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleGetRestartRequired(ctx, req, txlogClient)
	})

	// Tool: search_package
	searchPackageTool := mcp.NewTool("search_package",
		mcp.WithDescription("Busca quais assets têm um pacote específico instalado"),
		mcp.WithString("name",
			mcp.Description("Nome do pacote (ex: 'openssl', 'httpd')"),
			mcp.Required(),
		),
		mcp.WithString("version",
			mcp.Description("Versão específica do pacote"),
		),
		mcp.WithString("release",
			mcp.Description("Release específico do pacote"),
		),
	)

	s.AddTool(searchPackageTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleSearchPackage(ctx, req, txlogClient)
	})

	// Tool: get_transaction_details
	getTransactionDetailsTool := mcp.NewTool("get_transaction_details",
		mcp.WithDescription("Obtém os detalhes de uma transação específica, incluindo todos os pacotes alterados"),
		mcp.WithNumber("transaction_id",
			mcp.Description("ID da transação"),
			mcp.Required(),
		),
	)

	s.AddTool(getTransactionDetailsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleGetTransactionDetails(ctx, req, txlogClient)
	})
}

// handleListAssets handles the list_assets tool call.
func handleListAssets(_ context.Context, req mcp.CallToolRequest, txlogClient *client.Client) (*mcp.CallToolResult, error) {
	assets, err := txlogClient.ListAssets()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Erro ao listar assets: %v", err)), nil
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
		return mcp.NewToolResultText("Nenhum asset encontrado com os filtros especificados."), nil
	}

	return mcp.NewToolResultText(formatAssets(filtered)), nil
}

// handleGetAssetDetails handles the get_asset_details tool call.
func handleGetAssetDetails(_ context.Context, req mcp.CallToolRequest, txlogClient *client.Client) (*mcp.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	machineID := req.GetString("machine_id", "")

	if hostname == "" && machineID == "" {
		return mcp.NewToolResultError("É necessário fornecer hostname ou machine_id"), nil
	}

	var asset *client.Asset
	var err error

	if hostname != "" {
		asset, err = txlogClient.GetAssetByHostname(hostname)
	} else {
		// Search by machine_id
		assets, listErr := txlogClient.ListAssets()
		if listErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Erro ao buscar asset: %v", listErr)), nil
		}
		for _, a := range assets {
			if a.MachineID == machineID {
				asset = &a
				break
			}
		}
		if asset == nil {
			err = fmt.Errorf("asset não encontrado: %s", machineID)
		}
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Erro: %v", err)), nil
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
		return mcp.NewToolResultError(fmt.Sprintf("Erro ao listar transações: %v", err)), nil
	}

	if len(transactions) == 0 {
		return mcp.NewToolResultText("Nenhuma transação encontrada para este asset."), nil
	}

	return mcp.NewToolResultText(formatTransactions(transactions)), nil
}

// handleGetRestartRequired handles the get_restart_required tool call.
func handleGetRestartRequired(_ context.Context, req mcp.CallToolRequest, txlogClient *client.Client) (*mcp.CallToolResult, error) {
	assets, err := txlogClient.GetAssetsRequiringRestart()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Erro ao buscar assets: %v", err)), nil
	}

	if len(assets) == 0 {
		return mcp.NewToolResultText("✅ Nenhum asset precisa ser reiniciado no momento."), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("⚠️ **%d assets precisam ser reiniciados:**\n\n", len(assets)))
	for _, asset := range assets {
		sb.WriteString(fmt.Sprintf("- **%s** (%s)\n", asset.Hostname, asset.OS))
	}
	sb.WriteString("\nRecomendação: Agende uma janela de manutenção para reiniciar estes servidores.")

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
		return mcp.NewToolResultError(fmt.Sprintf("Erro ao buscar pacote: %v", err)), nil
	}

	if len(assets) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("Nenhum asset encontrado com o pacote %s instalado.", name)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**Pacote %s encontrado em %d assets:**\n\n", name, len(assets)))
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
		return mcp.NewToolResultError(fmt.Sprintf("Erro ao buscar detalhes da transação: %v", err)), nil
	}

	if len(items) == 0 {
		return mcp.NewToolResultText("Nenhum item encontrado para esta transação."), nil
	}

	return mcp.NewToolResultText(formatTransactionItems(items)), nil
}

// formatAssets formats a list of assets for display.
func formatAssets(assets []client.Asset) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**Total de assets: %d**\n\n", len(assets)))

	// Group by OS
	osCounts := make(map[string]int)
	for _, asset := range assets {
		osCounts[asset.OS]++
	}

	sb.WriteString("**Distribuição por OS:**\n")
	for os, count := range osCounts {
		sb.WriteString(fmt.Sprintf("- %s: %d\n", os, count))
	}

	sb.WriteString("\n**Lista de assets:**\n")
	for _, asset := range assets {
		status := "✅"
		if asset.NeedsRestarting {
			status = "⚠️ (precisa reiniciar)"
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
	sb.WriteString(fmt.Sprintf("- **Sistema Operacional:** %s\n", asset.OS))
	sb.WriteString(fmt.Sprintf("- **Versão do Agent:** %s\n", asset.AgentVersion))

	if asset.NeedsRestarting {
		sb.WriteString("- **Status:** ⚠️ Precisa reiniciar\n")
	} else {
		sb.WriteString("- **Status:** ✅ OK\n")
	}

	if len(transactions) > 0 {
		sb.WriteString("\n### Últimas Transações:\n")
		for _, t := range transactions {
			sb.WriteString(fmt.Sprintf("- [%s] %s - %d pacotes alterados\n",
				t.ExecutedAt, t.Cmdline, t.ItemsCount))
		}
	}

	return sb.String()
}

// formatTransactions formats transactions for display.
func formatTransactions(transactions []client.Transaction) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**Últimas %d transações:**\n\n", len(transactions)))

	for _, t := range transactions {
		sb.WriteString(fmt.Sprintf("### Transação #%d\n", t.ExternalID))
		sb.WriteString(fmt.Sprintf("- **Data:** %s\n", t.ExecutedAt))
		sb.WriteString(fmt.Sprintf("- **Usuário:** %s\n", t.Username))
		sb.WriteString(fmt.Sprintf("- **Comando:** `%s`\n", t.Cmdline))
		sb.WriteString(fmt.Sprintf("- **Pacotes alterados:** %d\n\n", t.ItemsCount))
	}

	return sb.String()
}

// formatTransactionItems formats transaction items for display.
func formatTransactionItems(items []client.TransactionItem) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**Detalhes da transação (%d pacotes):**\n\n", len(items)))

	// Group by action
	actions := make(map[string][]client.TransactionItem)
	for _, item := range items {
		actions[item.Action] = append(actions[item.Action], item)
	}

	actionLabels := map[string]string{
		"Install":   "📦 Instalados",
		"Update":    "🔄 Atualizados",
		"Remove":    "🗑️ Removidos",
		"Downgrade": "⬇️ Downgrade",
		"Reinstall": "♻️ Reinstalados",
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
