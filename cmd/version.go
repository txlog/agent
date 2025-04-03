package cmd

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const agentVersion = "1.2.3"

type ServerVersion struct {
	Version string `json:"version"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show agent and server version number",
	Run: func(cmd *cobra.Command, args []string) {
		serverVersion := "unknown"
		client := resty.New()
		client.SetAllowGetMethodPayload(true)

		var server ServerVersion
		_, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetResult(&server).
			Get(viper.GetString("server.url") + "/v1/version")

		if err == nil {
			serverVersion = server.Version
		}

		fmt.Println("Agent version " + agentVersion)
		fmt.Println("Server version " + serverVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
