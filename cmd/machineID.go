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

type MachineID struct {
	Hostname  string     `json:"hostname"`
	MachineID string     `json:"machine_id"`
	BeginTime *time.Time `json:"begin_time"`
}

// machineCmd represents the machine_id command
var machineCmd = &cobra.Command{
	Use:   "machine_id",
	Short: "List the machine_id of the given hostname",
	Long:  `This command lists the machine_id of a given hostname.`,
	Run: func(cmd *cobra.Command, args []string) {
		hostname, _ := cmd.Flags().GetString("hostname")

		machines, _, err := getMachineIDs(hostname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving machine IDs: %v\n", err)
			os.Exit(1)
		}

		if len(machines) > 0 {
			fmt.Println("| Hostname | Machine ID | Since |")
			fmt.Println("|----------|------------|-------|")
		}

		for _, exec := range machines {
			fmt.Printf("| %s | %s | %s |\n",
				exec.Hostname,
				exec.MachineID,
				exec.BeginTime.Format(time.RFC3339))
		}
	},
}

// getMachineIDs retrieves a list of machine IDs from the server.
func getMachineIDs(hostname string) ([]MachineID, int, error) {
	client := resty.New()
	client.SetAllowGetMethodPayload(true)

	var machines []MachineID
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"hostname": hostname,
		}).
		SetResult(&machines).
		Get(viper.GetString("server.url") + "/v1/machines/ids")

	if err != nil {
		return nil, 0, err
	}

	if response.StatusCode() != 200 {
		return nil, 0, fmt.Errorf("server returned status code %d: %s", response.StatusCode(), response.String())
	}

	return machines, len(machines), nil
}

func init() {
	rootCmd.AddCommand(machineCmd)
	hostname, _ := util.GetHostname()

	machineCmd.PersistentFlags().String("hostname", hostname, "The hostname, as returned by the 'hostname -s' command")
}
