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

// itemsCmd represents the items command
var itemsCmd = &cobra.Command{
	Use:   "items",
	Short: "List transactions items",
	Long:  `This command lists transactions items already compiled by txlog agent.`,
	Run: func(cmd *cobra.Command, args []string) {
		machineID, _ := cmd.Flags().GetString("machine_id")
		transactionID, _ := cmd.Flags().GetString("transaction_id")

		transaction, err := getItems(machineID, transactionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving transaction items: %v\n", err)
			os.Exit(1)
		}

		if transaction.Hostname == "" {
			fmt.Println("Error: Nothing found for Machine ID " + machineID + " and Transaction ID " + transactionID)
			os.Exit(1)
		} else if len(transaction.Items) == 0 {
			fmt.Println("")
			fmt.Println("* Hostname       : " + transaction.Hostname)
			fmt.Println("* Machine ID     : " + machineID)
			fmt.Println("* Transaction ID : " + transactionID)
			fmt.Println("* Items          : No items found")
			fmt.Println("")
		} else {
			fmt.Println("")
			fmt.Println("* Hostname         : " + transaction.Hostname)
			fmt.Println("* Machine ID       : " + machineID)
			fmt.Println("* Transaction ID   : " + transaction.TransactionID)
			fmt.Println("* Begin Time       : " + transaction.BeginTime.Format(time.RFC3339))
			fmt.Println("* End Time         : " + transaction.EndTime.Format(time.RFC3339))
			fmt.Println("* Actions          : " + transaction.Actions)
			fmt.Println("* Altered          : " + transaction.Altered)
			fmt.Println("* User             : " + transaction.User)
			fmt.Println("* Return Code      : " + transaction.ReturnCode)
			fmt.Println("* Release Version  : " + transaction.ReleaseVersion)
			fmt.Println("* Command Line     : " + transaction.CommandLine)
			fmt.Println("* Comment          : " + transaction.Comment)
			fmt.Println("* Scriptlet Output : " + transaction.ScriptletOutput)
			fmt.Printf("* Items            : %d\n", len(transaction.Items))
			fmt.Println("")

			fmt.Println("| Action | Package Name | Version | Release | Epoch | Arch | Repo |")
			fmt.Println("|--------|--------------|---------|---------|-------|------|------|")
		}

		for _, exec := range transaction.Items {
			fmt.Printf("| %s | %s | %s | %s | %s | %s | %s |\n",
				exec.Action,
				exec.Name,
				exec.Version,
				exec.Release,
				exec.Epoch,
				exec.Arch,
				exec.Repo)
		}
	},
}

// getItems retrieves the transaction items from the server
func getItems(machineID, transactionID string) (Transaction, error) {
	client := resty.New()
	client.SetAllowGetMethodPayload(true)
	if username := viper.GetString("server.username"); username != "" {
		client.SetBasicAuth(username, viper.GetString("server.password"))
	}

	var transaction Transaction
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"machine_id":     machineID,
			"transaction_id": transactionID,
		}).
		SetResult(&transaction).
		Get(viper.GetString("server.url") + "/v1/items")

	if err != nil {
		return Transaction{}, err
	}

	if response.StatusCode() != 200 {
		return Transaction{}, fmt.Errorf(
			"server returned status code %d: %s",
			response.StatusCode(),
			response.String(),
		)
	}

	return transaction, nil
}

func init() {
	rootCmd.AddCommand(itemsCmd)
	machineId, _ := util.GetMachineId()

	itemsCmd.PersistentFlags().String("machine_id", machineId, "The machine ID, as returned by the 'cat /etc/machine-id' command")
	itemsCmd.PersistentFlags().String("transaction_id", "", "The transaction ID to view. If not provided, the last transaction will be shown")
}
