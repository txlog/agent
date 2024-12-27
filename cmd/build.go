package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/txlog/agent/util"
)

// TransactionEntry represents a single transaction entry, as shown in the 'dnf history list' command.
type TransactionEntry struct {
	TransactionID string            `json:"transaction_id"`
	MachineID     string            `json:"machine_id"`
	Hostname      string            `json:"hostname"`
	CommandLine   string            `json:"command_line"`
	DateTime      string            `json:"date_time"`
	Actions       string            `json:"actions"`
	Altered       string            `json:"altered"`
	Details       TransactionDetail `json:"details"`
}

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
This command compiles all transactions of yum/dnf 'transaction' command, and
sends them to the server so they can be queried later.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(os.Stdout, "Compiling host identification...\n")
		machineId, _ := util.GetMachineId()
		hostname, _ := util.GetHostname()

		// * retrieves a list of all transactions saved on the server for this `machine-id`
		savedTransactions, _, err := getSavedTransactions(machineId, hostname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving saved transactions: %v\n", err)
			os.Exit(1)
		}

		// * compares the transaction lists to determine which transactions have not been sent to the server
		fmt.Fprintf(os.Stdout, "Compiling transaction data...\n")
		unsentTransactions, count, err := getUnsentTransactions(machineId, hostname, savedTransactions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving transactions: %v\n", err)
			os.Exit(1)
		}

		// * sends the unsent transactions to the server, one at a time, with data extracted from `sudo dnf history info ID`
		//    * The sending of the transaction and its details needs to be atomic

		fmt.Println(unsentTransactions)
		fmt.Println(count)

		// fmt.Fprintf(os.Stdout, "Done. %d transactions processed, %d transactions sent to server.\n", count, 0)

	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}

// getSavedTransactions retrieves a list of all transactions saved on the server for this `machine-id` and hostname.
func getSavedTransactions(machineId, hostname string) ([]TransactionEntry, int, error) {
	client := resty.New()

	var transactions []TransactionEntry
	response, err := client.R().
		SetQueryParams(map[string]string{
			"machine_id": "eq." + machineId,
			"hostname":   "eq." + hostname,
		}).
		SetHeader("Accept", "application/json").
		SetAuthToken(viper.GetString("postgrest.jwt_token")).
		SetResult(&transactions).
		Get(viper.GetString("postgrest.url") + "/transactions")

	if err != nil {
		return nil, 0, err
	}

	if response.StatusCode() != 200 {
		return nil, 0, fmt.Errorf("server returned status code %d", response.StatusCode())
	}

	return transactions, len(transactions), nil
}

// getUnsentTransactionItems retrieves the details of all transaction entries which have not been sent to the server.
func getUnsentTransactions(machineId, hostname string, savedTransactions []TransactionEntry) (string, int, error) {
	out, err := exec.Command(util.PackageBinary(), "history", "--reverse", "list").Output()
	if err != nil {
		return "", 0, err
	}

	output := string(out)
	lines := strings.Split(output, "\n")
	lines = lines[2:]

	re := regexp.MustCompile(`\s*(\d+)\s*\|\s*(.+?)\s*\|\s*(.+?)\s*\|\s*(.+?)\s*\|\s*(.+?)\s*`)
	var entries []TransactionEntry
	count := 0

	for _, line := range lines {
		if re.MatchString(line) {
			matches := re.FindStringSubmatch(line)
			transactionID := strings.TrimSpace(matches[1])
			fmt.Fprintf(os.Stdout, "  #%s... ", transactionID)

			exists := false
			for _, t := range savedTransactions {
				if t.TransactionID == transactionID {
					exists = true
					break
				}
			}

			if !exists {
				details, err := getTransactionItems(transactionID)
				if err != nil {
					return "", 0, err
				}

				entry := TransactionEntry{
					TransactionID: transactionID,
					MachineID:     machineId,
					Hostname:      hostname,
					CommandLine:   strings.TrimSpace(matches[2]),
					DateTime:      strings.TrimSpace(matches[3]),
					Actions:       strings.TrimSpace(matches[4]),
					Altered:       strings.TrimSpace(matches[5]),
					Details:       details,
				}
				entries = append(entries, entry)
				count++
				fmt.Fprintf(os.Stdout, "done.\n")
			} else {
				fmt.Fprintf(os.Stdout, "already sent. Skipping.\n")
			}
		}
	}

	jsonData, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		fmt.Println("Erro ao converter para JSON:", err)
		return "", 0, err
	}

	return string(jsonData), count, nil
}

func getTransactionItems(transaction_id string) (TransactionDetail, error) {
	out, err := exec.Command("dnf", "history", "info", transaction_id).Output()
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
		if reTransaction.MatchString(line) {
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
		} else if rePackage.MatchString(line) {
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
		}
	}

	transaction.PackagesAltered = packages
	transaction.ScriptletOutput = scriptletOutput

	return transaction, nil
}
