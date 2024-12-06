package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
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
		// * Compilação de transações (txlog build)
		//   * Agente obtém `/etc/machine-id` e `hostname`
		//   * Agente obtém lista de todas as transactions com `sudo dnf history --reverse list`
		transactions, _ := getTransactions()
		//   * Agente obtém lista de todas as transactions salvas no servidor, para este `machine-id`
		//   * Agente compara as listas de transações, para ver quais não foram enviadas ao servidor
		//   * Agente envia transações que não foram enviadas para o servidor, uma de cada vez, com dados extraídos de `sudo dnf history info ID`
		fmt.Println(transactions)

	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}

func getTransactions() (string, error) {
	out, err := exec.Command(util.PackageBinary(), "history", "--reverse", "list").Output()
	if err != nil {
		return "", err
	}

	output := string(out)
	lines := strings.Split(output, "\n")
	lines = lines[2:]

	re := regexp.MustCompile(`\s*(\d+)\s*\|\s*(.+?)\s*\|\s*(.+?)\s*\|\s*(.+?)\s*\|\s*(.+?)\s*`)
	var entries []TransactionEntry
	for _, line := range lines {
		if re.MatchString(line) {
			matches := re.FindStringSubmatch(line)
			details, err := getTransactionItems(strings.TrimSpace(matches[1]))
			machineId, _ := util.GetMachineId()
			hostname, _ := util.GetHostname()

			if err != nil {
				return "", err
			}

			entry := TransactionEntry{
				TransactionID: strings.TrimSpace(matches[1]),
				MachineID:     machineId,
				Hostname:      hostname,
				CommandLine:   strings.TrimSpace(matches[2]),
				DateTime:      strings.TrimSpace(matches[3]),
				Actions:       strings.TrimSpace(matches[4]),
				Altered:       strings.TrimSpace(matches[5]),
				Details:       details,
			}
			entries = append(entries, entry)
		}
	}

	jsonData, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		fmt.Println("Erro ao converter para JSON:", err)
		return "", err
	}

	return string(jsonData), nil
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
