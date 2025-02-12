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

type Execution struct {
	ExecutionID           string     `json:"execution_id,omitempty"`
	MachineID             string     `json:"machine_id"`
	Hostname              string     `json:"hostname"`
	ExecutedAt            *time.Time `json:"executed_at"`
	Success               bool       `json:"success"`
	Details               string     `json:"details,omitempty"`
	TransactionsProcessed int        `json:"transactions_processed,omitempty"`
	TransactionsSent      int        `json:"transactions_sent,omitempty"`
}

// executionsCmd represents the executions command
var executionsCmd = &cobra.Command{
	Use:   "executions",
	Short: "List build executions",
	Long: `
This command lists build executions already made by txlog. If no parameter is passed,
it will display the executions of the current host. However, when a parameter is passed,
you can consult the executions of other hosts.`,
	Run: func(cmd *cobra.Command, args []string) {
		machineID, _ := cmd.Flags().GetString("machine_id")
		success, _ := cmd.Flags().GetString("success")

		savedExecutions, _, err := getSavedExecutions(machineID, success)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving saved executions: %v\n", err)
			os.Exit(1)
		}

		if len(savedExecutions) > 0 {
			fmt.Println("| ID | Host | Success | Executed | Processed | Sent | Details |")
			fmt.Println("|-----|------|---------|----------|-----------|------|---------|")
		}

		for _, exec := range savedExecutions {
			fmt.Printf("| %s | %s | %v | %v | %v | %v | %v |\n",
				exec.ExecutionID,
				exec.Hostname,
				exec.Success,
				exec.ExecutedAt.Format(time.RFC3339),
				exec.TransactionsProcessed,
				exec.TransactionsSent,
				exec.Details)
		}
	},
}

// getSavedExecutions retrieves a list of saved executions from the server.
func getSavedExecutions(machineId, success string) ([]Execution, int, error) {
	client := resty.New()
	client.SetAllowGetMethodPayload(true)

	var executions []Execution
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"machine_id": machineId,
			"success":    success,
		}).
		SetResult(&executions).
		Get(viper.GetString("server.url") + "/v1/execution")

	if err != nil {
		return nil, 0, err
	}

	if response.StatusCode() != 200 {
		return nil, 0, fmt.Errorf("server returned status code %d: %s", response.StatusCode(), response.String())
	}

	return executions, len(executions), nil
}

func init() {
	rootCmd.AddCommand(executionsCmd)
	machineId, _ := util.GetMachineId()

	executionsCmd.PersistentFlags().String("machine_id", machineId, "The machine ID, as returned by the 'cat /etc/machine-id' command")
	executionsCmd.PersistentFlags().String("success", "", "Filter by execution success status (true/false)")
}
