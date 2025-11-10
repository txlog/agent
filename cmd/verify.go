package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/txlog/agent/util"
)

// ServerTransaction represents a transaction as stored on the server
type ServerTransaction struct {
	TransactionID   string    `json:"transaction_id"`
	Hostname        string    `json:"hostname"`
	BeginTime       string    `json:"begin_time"`
	EndTime         string    `json:"end_time"`
	Actions         string    `json:"actions"`
	Altered         string    `json:"altered"`
	User            string    `json:"user"`
	ReturnCode      string    `json:"return_code"`
	ReleaseVersion  string    `json:"release_version"`
	CommandLine     string    `json:"command_line"`
	Comment         string    `json:"comment"`
	ScriptletOutput string    `json:"scriptlet_output"`
	Items           []Package `json:"items"`
}

// VerificationResult holds the results of the verification process
type VerificationResult struct {
	TotalLocalTransactions  int
	TotalServerTransactions int
	MissingOnServer         []string
	WithMissingItems        []string
	WithExtraItems          []string
	FullyVerified           int
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify data integrity between local DNF history and server",
	Long: `
This command verifies that all local DNF transaction data has been properly
replicated to the server. It checks:
  - Transactions that exist locally but not on the server
  - Transaction items (packages) that may be missing or different on the server`,
	Run: func(cmd *cobra.Command, args []string) {
		machineId, err := util.GetMachineId()
		if err != nil {
			color.Red("Error getting machine ID: %v", err)
			os.Exit(1)
		}

		hostname, err := util.GetHostname()
		if err != nil {
			color.Red("Error getting hostname: %v", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "Verifying data integrity for %s\n", color.CyanString(hostname))
		fmt.Fprintf(os.Stdout, "Machine ID: %s\n\n", color.CyanString(machineId))

		result, err := verifyDataIntegrity(machineId, hostname)
		if err != nil {
			color.Red("Error during verification: %v", err)
			os.Exit(1)
		}

		printVerificationResults(result)

		// Exit with error code if there are any issues
		if len(result.MissingOnServer) > 0 || len(result.WithMissingItems) > 0 || len(result.WithExtraItems) > 0 {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}

// verifyDataIntegrity performs the complete data integrity verification
func verifyDataIntegrity(machineId, hostname string) (*VerificationResult, error) {
	result := &VerificationResult{
		MissingOnServer:  make([]string, 0),
		WithMissingItems: make([]string, 0),
		WithExtraItems:   make([]string, 0),
	}

	// Get local transactions
	fmt.Fprintf(os.Stdout, "Reading local transaction history...\n")
	localTransactions, err := getLocalTransactionIDs()
	if err != nil {
		return nil, fmt.Errorf("error reading local transactions: %w", err)
	}
	result.TotalLocalTransactions = len(localTransactions)
	fmt.Fprintf(os.Stdout, "Found %s local transactions\n", color.YellowString("%d", len(localTransactions)))

	// Get server transactions
	fmt.Fprintf(os.Stdout, "Retrieving server transactions...\n")
	serverTransactionIDs, _, err := getSavedTransactions(machineId, hostname)
	if err != nil {
		return nil, fmt.Errorf("error retrieving server transactions: %w", err)
	}
	result.TotalServerTransactions = len(serverTransactionIDs)
	fmt.Fprintf(os.Stdout, "Found %s transactions on server\n\n", color.YellowString("%d", len(serverTransactionIDs)))

	// Convert server IDs to map for quick lookup
	serverTransactionsMap := make(map[int]bool)
	for _, id := range serverTransactionIDs {
		serverTransactionsMap[id] = true
	}

	// Check for missing transactions on server
	fmt.Fprintf(os.Stdout, "Checking for missing transactions...\n")
	for _, localID := range localTransactions {
		if !serverTransactionsMap[localID] {
			result.MissingOnServer = append(result.MissingOnServer, fmt.Sprintf("%d", localID))
			color.Red("  ✗ Transaction #%d exists locally but not on server", localID)
		}
	}

	if len(result.MissingOnServer) == 0 {
		color.Green("  ✓ All local transactions exist on server")
	}
	fmt.Fprintln(os.Stdout)

	// Check transaction items for each transaction on server
	fmt.Fprintf(os.Stdout, "Verifying transaction items integrity...\n")
	for _, serverID := range serverTransactionIDs {
		// Skip verification if transaction doesn't exist locally
		found := false
		for _, localID := range localTransactions {
			if localID == serverID {
				found = true
				break
			}
		}
		if !found {
			continue
		}

		// Get local transaction details
		localDetails, err := getTransactionItems(fmt.Sprintf("%d", serverID))
		if err != nil {
			color.Yellow("  ⚠ Warning: Could not get local details for transaction #%d: %v", serverID, err)
			continue
		}

		// Get server transaction details using the /v1/items endpoint
		serverDetails, err := getServerTransactionItems(machineId, fmt.Sprintf("%d", serverID))
		if err != nil {
			color.Yellow("  ⚠ Warning: Could not get server details for transaction #%d: %v", serverID, err)
			continue
		}

		// Compare items
		missing, extra := compareTransactionItems(localDetails.PackagesAltered, serverDetails.Items)

		if len(missing) > 0 {
			result.WithMissingItems = append(result.WithMissingItems, fmt.Sprintf("%d", serverID))
			color.Red("  ✗ Transaction #%d is missing %d package(s) on server", serverID, len(missing))
			for _, pkg := range missing {
				fmt.Fprintf(os.Stdout, "    - %s %s-%s.%s (%s)\n", pkg.Action, pkg.Name, pkg.Version, pkg.Arch, pkg.Repo)
			}
		}

		if len(extra) > 0 {
			result.WithExtraItems = append(result.WithExtraItems, fmt.Sprintf("%d", serverID))
			color.Yellow("  ⚠ Transaction #%d has %d extra package(s) on server", serverID, len(extra))
			for _, pkg := range extra {
				fmt.Fprintf(os.Stdout, "    + %s %s-%s.%s\n", pkg.Action, pkg.Name, pkg.Version, pkg.Arch)
			}
		}

		if len(missing) == 0 && len(extra) == 0 {
			result.FullyVerified++
		}
	}

	if result.FullyVerified == len(serverTransactionIDs) {
		color.Green("  ✓ All transaction items verified successfully")
	}
	fmt.Fprintln(os.Stdout)

	return result, nil
}

// getLocalTransactionIDs retrieves all transaction IDs from local DNF history
func getLocalTransactionIDs() ([]int, error) {
	out, err := exec.Command(util.PackageBinary(), "history", "list").Output()
	if err != nil {
		return nil, err
	}

	output := string(out)
	lines := strings.Split(output, "\n")
	if len(lines) < 3 {
		return nil, fmt.Errorf("unexpected dnf history output format")
	}

	// Skip header lines
	lines = lines[2:]

	re := regexp.MustCompile(`^\s*(\d+)\s*\|`)
	transactionIDs := make([]int, 0)

	for _, line := range lines {
		if re.MatchString(line) {
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				id, err := strconv.Atoi(strings.TrimSpace(matches[1]))
				if err == nil {
					transactionIDs = append(transactionIDs, id)
				}
			}
		}
	}

	return transactionIDs, nil
}

// getServerTransactionItems retrieves detailed information about a specific transaction from the server
// using the /v1/items endpoint
func getServerTransactionItems(machineId, transactionID string) (*ServerTransaction, error) {
	client := resty.New()

	var transaction ServerTransaction
	request := client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParam("machine_id", machineId).
		SetQueryParam("transaction_id", transactionID).
		SetResult(&transaction)

	util.SetAuthentication(request)

	response, err := request.Get(viper.GetString("server.url") + "/v1/items")

	if err != nil {
		return nil, err
	}

	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("server returned status code %d: %s", response.StatusCode(), response.String())
	}

	return &transaction, nil
}

// compareTransactionItems compares local and server package lists
// Returns: (missing packages, extra packages)
func compareTransactionItems(localPackages, serverPackages []Package) ([]Package, []Package) {
	missing := make([]Package, 0)
	extra := make([]Package, 0)

	// Create a map of server packages for quick lookup
	serverPkgMap := make(map[string]Package)
	for _, pkg := range serverPackages {
		key := fmt.Sprintf("%s|%s|%s|%s|%s|%s", pkg.Action, pkg.Name, pkg.Version, pkg.Release, pkg.Epoch, pkg.Arch)
		serverPkgMap[key] = pkg
	}

	// Create a map of local packages
	localPkgMap := make(map[string]Package)
	for _, pkg := range localPackages {
		key := fmt.Sprintf("%s|%s|%s|%s|%s|%s", pkg.Action, pkg.Name, pkg.Version, pkg.Release, pkg.Epoch, pkg.Arch)
		localPkgMap[key] = pkg
	}

	// Check for missing packages (in local but not in server)
	for key, pkg := range localPkgMap {
		if _, exists := serverPkgMap[key]; !exists {
			missing = append(missing, pkg)
		}
	}

	// Check for extra packages (in server but not in local)
	for key, pkg := range serverPkgMap {
		if _, exists := localPkgMap[key]; !exists {
			extra = append(extra, pkg)
		}
	}

	return missing, extra
}

// printVerificationResults prints a summary of the verification results
func printVerificationResults(result *VerificationResult) {
	fmt.Fprintln(os.Stdout, strings.Repeat("=", 60))
	fmt.Fprintln(os.Stdout, "VERIFICATION SUMMARY")
	fmt.Fprintln(os.Stdout, strings.Repeat("=", 60))
	fmt.Fprintf(os.Stdout, "Total local transactions:  %s\n", color.CyanString("%d", result.TotalLocalTransactions))
	fmt.Fprintf(os.Stdout, "Total server transactions: %s\n", color.CyanString("%d", result.TotalServerTransactions))
	fmt.Fprintf(os.Stdout, "Fully verified:            %s\n", color.GreenString("%d", result.FullyVerified))
	fmt.Fprintln(os.Stdout, strings.Repeat("-", 60))

	if len(result.MissingOnServer) > 0 {
		fmt.Fprintf(os.Stdout, "Missing on server:         %s\n", color.RedString("%d", len(result.MissingOnServer)))
	} else {
		fmt.Fprintf(os.Stdout, "Missing on server:         %s\n", color.GreenString("0"))
	}

	if len(result.WithMissingItems) > 0 {
		fmt.Fprintf(os.Stdout, "With missing items:        %s\n", color.RedString("%d", len(result.WithMissingItems)))
	} else {
		fmt.Fprintf(os.Stdout, "With missing items:        %s\n", color.GreenString("0"))
	}

	if len(result.WithExtraItems) > 0 {
		fmt.Fprintf(os.Stdout, "With extra items:          %s\n", color.YellowString("%d", len(result.WithExtraItems)))
	} else {
		fmt.Fprintf(os.Stdout, "With extra items:          %s\n", color.GreenString("0"))
	}

	fmt.Fprintln(os.Stdout, strings.Repeat("=", 60))

	if len(result.MissingOnServer) == 0 && len(result.WithMissingItems) == 0 && len(result.WithExtraItems) == 0 {
		color.Green("\n✓ Data integrity verified successfully!")
		fmt.Fprintln(os.Stdout, "All local transactions and items are properly replicated on the server.")
	} else {
		color.Red("\n✗ Data integrity issues detected!")
		fmt.Fprintln(os.Stdout, "To fix these issues:")
		fmt.Fprintln(os.Stdout, "  1. Remove the hostname data from the server")
		fmt.Fprintln(os.Stdout, "  2. Run 'txlog build' to synchronize all data again")
	}
	fmt.Fprintln(os.Stdout)
}
