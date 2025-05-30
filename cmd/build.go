package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/txlog/agent/util"
)

// TransactionDetail represents a detailed transaction entry, as shown in the 'dnf history info' command.
type TransactionDetail struct {
	TransactionID   string    `json:"transaction_id"`
	BeginTime       string    `json:"begin_time"`
	BeginRPMDB      string    `json:"begin_rpmdb"`
	EndTime         string    `json:"end_time"`
	EndRPMDB        string    `json:"end_rpmdb"`
	User            string    `json:"user"`
	ReturnCode      string    `json:"return_code"`
	Releasever      string    `json:"releasever"`
	CommandLine     string    `json:"command_line"`
	Comment         string    `json:"comment"`
	PackagesAltered []Package `json:"packages_altered"`
	ScriptletOutput []string  `json:"scriptlet_output"`
}

// Package represents a single package in a transaction entry.
type Package struct {
	Action   string `json:"action"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Release  string `json:"release"`
	Epoch    string `json:"epoch"`
	Arch     string `json:"arch"`
	Repo     string `json:"repo"`
	FromRepo string `json:"from_repo,omitempty"`
}

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Compile transaction info",
	Long: `
This command compiles all transactions listed on 'dnf history'
command, and sends them to the server so they can be queried later.`,
	Run: func(cmd *cobra.Command, args []string) {
		machineId, _ := util.GetMachineId()
		hostname, _ := util.GetHostname()
		fmt.Fprintf(os.Stdout, "Compiling host identification for %s\n", hostname)

		// * retrieves a list of all transactions saved on the server for this `machine-id`
		fmt.Fprintf(os.Stdout, "Retrieving saved transactions\n")
		savedTransactions, _, err := getSavedTransactions(machineId, hostname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving saved transactions: %v\n", err)
			saveExecution(false, machineId, hostname, err.Error(), 0, 0)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "Compiling transaction data\n")
		// * compares the transaction lists to determine which transactions have not been sent to the server
		// * sends the unsent transactions to the server, one at a time, with data extracted from `sudo dnf history info ID`
		//    * The sending of the transaction and its details needs to be atomic
		entriesProcessed, entriesSent, err := saveUnsentTransactions(machineId, hostname, savedTransactions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving transactions: %v\n", err)
			saveExecution(false, machineId, hostname, err.Error(), 0, 0)
			os.Exit(1)
		}

		saveExecution(true, machineId, hostname, "", entriesProcessed, entriesSent)
		fmt.Fprintf(os.Stdout, "Done. %d transactions processed, %d transactions sent to server.\n", entriesProcessed, entriesSent)

	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}

// getSavedTransactions retrieves transaction IDs from a remote server for a given machine ID and hostname.
// It makes an HTTP GET request to the configured server endpoint with the machine ID and hostname in the request body.
//
// Parameters:
//   - machineId: string containing the unique identifier for the machine
//   - hostname: string containing the hostname of the machine
//
// Returns:
//   - []int: slice of transaction IDs retrieved from the server
//   - int: number of transactions retrieved
//   - error: nil if successful, otherwise contains error information
//     Possible errors include network failures or non-200 HTTP status codes
func getSavedTransactions(machineId, hostname string) ([]int, int, error) {
	client := resty.New()
	client.SetAllowGetMethodPayload(true)

	var transactions []int
	request := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"machine_id": machineId,
			"hostname":   hostname,
		}).
		SetResult(&transactions)

	if username := viper.GetString("server.username"); username != "" {
		request.SetBasicAuth(username, viper.GetString("server.password"))
	}

	response, err := request.Get(viper.GetString("server.url") + "/v1/transactions/ids")

	if err != nil {
		return nil, 0, err
	}

	if response.StatusCode() != 200 {
		return nil, 0, fmt.Errorf("server returned status code %d: %s", response.StatusCode(), response.String())
	}

	return transactions, len(transactions), nil
}

// saveUnsentTransactions processes DNF transaction history and sends unsent transactions to a remote server.
// It takes the machine ID, hostname, and a slice of previously saved transaction IDs as input.
//
// The function performs the following steps:
// 1. Retrieves the DNF transaction history in reverse order
// 2. Parses each transaction entry
// 3. For transactions not previously saved:
//   - Gets detailed transaction information
//   - Sends the transaction data to the configured server endpoint
//
// Parameters:
//   - machineId: string identifier for the machine
//   - hostname: system hostname
//   - savedTransactions: slice of previously processed transaction IDs to avoid duplication
//
// Returns:
//   - int: total number of entries processed
//   - int: number of new entries sent to server
//   - error: any error encountered during execution
func saveUnsentTransactions(machineId, hostname string, savedTransactions []int) (int, int, error) {
	out, err := exec.Command(util.PackageBinary(), "history", "--reverse", "list").Output()
	if err != nil {
		return 0, 0, err
	}

	output := string(out)
	lines := strings.Split(output, "\n")
	lines = lines[2:]

	re := regexp.MustCompile(`\s*(\d+)\s*\|\s*(.+?)\s*\|\s*(.+?)\s*\|\s*(.+?)\s*\|\s*(.+?)\s*`)
	entriesProcessed := 0
	entriesSent := 0

	client := resty.New()

	for _, line := range lines {
		if re.MatchString(line) {
			matches := re.FindStringSubmatch(line)
			transactionID := strings.TrimSpace(matches[1])

			exists := false
			for _, t := range savedTransactions {
				if fmt.Sprintf("%d", t) == transactionID {
					exists = true
					break
				}
			}

			if !exists {
				details, err := getTransactionItems(transactionID)
				if err != nil {
					return 0, 0, err
				}

				request := client.R().
					SetHeader("Content-Type", "application/json").
					SetBody(map[string]interface{}{
						"transaction_id":   transactionID,
						"machine_id":       machineId,
						"hostname":         hostname,
						"begin_time":       details.BeginTime,
						"end_time":         details.EndTime,
						"actions":          strings.TrimSpace(matches[4]),
						"altered":          strings.TrimSpace(matches[5]),
						"user":             details.User,
						"return_code":      details.ReturnCode,
						"release_version":  details.Releasever,
						"command_line":     details.CommandLine,
						"comment":          details.Comment,
						"scriptlet_output": strings.Join(details.ScriptletOutput, "\n"),
						"items":            details.PackagesAltered,
					})

				if username := viper.GetString("server.username"); username != "" {
					request.SetBasicAuth(username, viper.GetString("server.password"))
				}

				response, err := request.Post(viper.GetString("server.url") + "/v1/transactions")

				if err != nil {
					return 0, 0, err
				}

				if response.StatusCode() != 200 {
					return 0, 0, fmt.Errorf("server returned status code %d: %s", response.StatusCode(), response.String())
				}

				entriesSent++
				fmt.Fprintf(os.Stdout, "Transaction #%s sent.\n", transactionID)
			}
			entriesProcessed++
		}
	}

	return entriesProcessed, entriesSent, nil
}

func getTransactionItems(transaction_id string) (TransactionDetail, error) {
	validInput := regexp.MustCompile(`^[a-zA-Z0-9_\-\./\\]+$`)
	if !validInput.MatchString(transaction_id) {
		return TransactionDetail{}, fmt.Errorf("invalid input")
	}
	out, err := exec.Command(util.PackageBinary(), "history", "info", transaction_id).Output()
	if err != nil {
		return TransactionDetail{}, err
	}

	output := string(out)
	lines := strings.Split(output, "\n")

	// Use regex to extract data
	reTransaction := regexp.MustCompile(`^(.+?)\s*:\s*(.+)$`)
	rePackage := regexp.MustCompile(`^\s+(\w+)\s+(.+?)\s+@(.+)$`)
	rePackageUpgraded := regexp.MustCompile(`^\s+(\w+)\s+(.+?)\s+@@(.+)$`)

	var transaction TransactionDetail
	var packages []Package
	var scriptletOutput []string
	for _, line := range lines {
		if rePackage.MatchString(line) {
			matches := rePackage.FindStringSubmatch(line)
			name, version, release, epoch, arch := util.SplitPackageName(strings.TrimSpace(matches[2]))
			pkg := Package{
				Action:  strings.TrimSpace(matches[1]),
				Name:    name,
				Version: version,
				Release: release,
				Epoch:   epoch,
				Arch:    arch,
				Repo:    strings.TrimSpace(matches[3]),
			}
			packages = append(packages, pkg)
		} else if rePackageUpgraded.MatchString(line) {
			matches := rePackageUpgraded.FindStringSubmatch(line)
			name, version, release, epoch, arch := util.SplitPackageName(strings.TrimSpace(matches[2]))
			pkg := Package{
				Action:   strings.TrimSpace(matches[1]),
				Name:     name,
				Version:  version,
				Release:  release,
				Epoch:    epoch,
				Arch:     arch,
				FromRepo: strings.TrimSpace(matches[3]),
			}
			packages = append(packages, pkg)
		} else if strings.HasPrefix(line, "  ") {
			scriptletOutput = append(scriptletOutput, strings.TrimSpace(line))
		} else if reTransaction.MatchString(line) {
			matches := reTransaction.FindStringSubmatch(line)
			key := strings.TrimSpace(matches[1])
			value := strings.TrimSpace(matches[2])
			switch key {
			case "Transaction ID":
				transaction.TransactionID = value
			case "Begin time":
				beginTime, err := util.DateConversion(value)
				if err != nil {
					return TransactionDetail{}, err
				}
				transaction.BeginTime = beginTime
			case "Begin rpmdb":
				transaction.BeginRPMDB = value
			case "End time":
				endTime, err := util.DateConversion(strings.Split(value, " (")[0])
				if err != nil {
					return TransactionDetail{}, err
				}
				transaction.EndTime = endTime
			case "End rpmdb":
				transaction.EndRPMDB = value
			case "User":
				transaction.User = value
			case "Return-Code":
				transaction.ReturnCode = value
			case "Releasever":
				transaction.Releasever = value
			case "Command Line":
				transaction.CommandLine = value
			case "Comment":
				transaction.Comment = value
			}
		}
	}

	transaction.PackagesAltered = packages
	transaction.ScriptletOutput = scriptletOutput

	return transaction, nil
}

// saveExecution sends the execution details to the server.
func saveExecution(success bool, machineId, hostname, details string, processed, sent int) error {
	err := util.ParseOSRelease()
	if err != nil {
		return fmt.Errorf("error while reading /etc/os-release file: %v", err)
	}

	// * retrieves the server version
	serverVersion := GetServerVersion()

	body := map[string]interface{}{
		"machine_id":             machineId,
		"hostname":               hostname,
		"executed_at":            time.Now().Format("2006-01-02T15:04:05Z07:00"),
		"details":                details,
		"success":                success,
		"transactions_processed": processed,
		"transactions_sent":      sent,
		"agent_version":          agentVersion,
		"os":                     util.Release.PrettyName,
	}

	if serverVersion != "unknown" && serverVersion >= "1.8.0" {
		needsRestarting, reason := util.NeedsRestarting()
		body["needs_restarting"] = needsRestarting
		body["restarting_reason"] = reason
	}

	client := resty.New()
	request := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body)
	if username := viper.GetString("server.username"); username != "" {
		request.SetBasicAuth(username, viper.GetString("server.password"))
	}

	response, err := request.Post(viper.GetString("server.url") + "/v1/executions")

	if err != nil {
		return err
	}

	if response.StatusCode() != 200 {
		return fmt.Errorf("server returned status code %d: %s", response.StatusCode(), response.String())
	}

	return nil
}
