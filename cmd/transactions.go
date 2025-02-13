package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/txlog/agent/util"
)

type Transaction struct {
	TransactionID   string            `json:"transaction_id"`
	MachineID       string            `json:"machine_id,omitempty"`
	Hostname        string            `json:"hostname"`
	BeginTime       *time.Time        `json:"begin_time"`
	EndTime         *time.Time        `json:"end_time"`
	Actions         string            `json:"actions"`
	Altered         string            `json:"altered"`
	User            string            `json:"user"`
	ReturnCode      string            `json:"return_code"`
	ReleaseVersion  string            `json:"release_version"`
	CommandLine     string            `json:"command_line"`
	Comment         string            `json:"comment"`
	ScriptletOutput string            `json:"scriptlet_output"`
	Items           []TransactionItem `json:"items"`
}

type TransactionItem struct {
	Action   string `json:"action"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Release  string `json:"release"`
	Epoch    string `json:"epoch"`
	Arch     string `json:"arch"`
	Repo     string `json:"repo"`
	FromRepo string `json:"from_repo"`
}

// transactionsCmd represents the transactions command
var transactionsCmd = &cobra.Command{
	Use:   "transactions",
	Short: "List compiled transactions",
	Long: `
This command lists transactions already executed by yum/dnf. By default, it
displays transactions from the current host. However, when a parameter is passed,
you can consult the transactions of other hosts.`,
	Run: func(cmd *cobra.Command, args []string) {
		machineID, _ := cmd.Flags().GetString("machine_id")

		transactions, _, err := getTransactions(machineID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving transactions: %v\n", err)
			os.Exit(1)
		}

		if len(transactions) > 0 {
			fmt.Println("")
			fmt.Println("* Hostname    : " + transactions[0].Hostname)
			fmt.Println("* Machine ID  : " + machineID)
			fmt.Printf("* Transactions: %d\n", len(transactions))
			fmt.Println("")

			fmt.Println("| Transaction ID | Start | Actions | Altered | User | Return Code | Command Line |")
			fmt.Println("|----------------|-------|---------|---------|------|-------------|--------------|")
		} else {
			fmt.Println("* Machine ID  : " + machineID)
			fmt.Println("* Transactions: 0")
		}

		for _, exec := range transactions {
			fmt.Printf("| %s | %s | %s | %s | %s | %s | %s |\n",
				exec.TransactionID,
				exec.BeginTime.Format(time.RFC3339),
				exec.Actions,
				exec.Altered,
				exec.User,
				exec.ReturnCode,
				exec.CommandLine)
		}
	},
}

// getTransactions retrieves a list of transactions from the server.
func getTransactions(machineID string) ([]Transaction, int, error) {
	client := resty.New()
	client.SetAllowGetMethodPayload(true)

	var transactions []Transaction
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"machine_id": machineID,
		}).
		SetResult(&transactions).
		Get(viper.GetString("server.url") + "/v1/transactions")

	if err != nil {
		return nil, 0, err
	}

	if response.StatusCode() != 200 {
		return nil, 0, fmt.Errorf("server returned status code %d: %s", response.StatusCode(), response.String())
	}

	return transactions, len(transactions), nil
}

func init() {
	rootCmd.AddCommand(transactionsCmd)
	machineId, _ := util.GetMachineId()

	transactionsCmd.PersistentFlags().String("machine_id", machineId, "The machine ID, as returned by the 'cat /etc/machine-id' command")
}
