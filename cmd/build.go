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
This command compiles all transactions of yum/dnf 'transaction' command, and
sends them to the server so they can be queried later.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(os.Stdout, "Compiling host identification\n")
		machineId, _ := util.GetMachineId()
		hostname, _ := util.GetHostname()

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

// getSavedTransactions retrieves the list of transactions saved on the server for the given machine ID.
func getSavedTransactions(machineId, hostname string) ([]int, int, error) {
	client := resty.New()
	client.SetAllowGetMethodPayload(true)

	var transactions []int
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"machine_id": machineId,
			"hostname":   hostname,
		}).
		SetResult(&transactions).
		Get(viper.GetString("server.url") + "/v1/transaction_id")

	if err != nil {
		return nil, 0, err
	}

	if response.StatusCode() != 200 {
		return nil, 0, fmt.Errorf("server returned status code %d: %s", response.StatusCode(), response.String())
	}

	return transactions, len(transactions), nil
}

// saveUnsentTransactionItems sends the transaction details to the server.
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
			fmt.Fprintf(os.Stdout, "Transaction #%s", transactionID)

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

				response, err := client.R().
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
					}).
					Post(viper.GetString("server.url") + "/v1/transaction")

				if err != nil {
					return 0, 0, err
				}

				if response.StatusCode() != 200 {
					return 0, 0, fmt.Errorf("server returned status code %d: %s", response.StatusCode(), response.String())
				}

				entriesSent++
				fmt.Fprintf(os.Stdout, " sent.\n")
			} else {
				fmt.Fprintf(os.Stdout, " already sent. Skipping.\n")
			}
			entriesProcessed++
		}
	}

	return entriesProcessed, entriesSent, nil
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

// saveExecution sends the execution details to the server.
func saveExecution(success bool, machineId, hostname, details string, processed, sent int) error {
	response, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"machine_id":             machineId,
			"hostname":               hostname,
			"executed_at":            time.Now().Format("2006-01-02T15:04:05Z07:00"),
			"details":                details,
			"success":                success,
			"transactions_processed": processed,
			"transactions_sent":      sent,
		}).
		Post(viper.GetString("server.url") + "/v1/execution")

	if err != nil {
		return err
	}

	if response.StatusCode() != 200 {
		return fmt.Errorf("server returned status code %d: %s", response.StatusCode(), response.String())
	}

	return nil
}
