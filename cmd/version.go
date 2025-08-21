package cmd

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const agentVersion = "1.6.1"

type ServerVersion struct {
	Version string `json:"version"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show agent and server version number",
	Run: func(cmd *cobra.Command, args []string) {
		serverVersion := GetServerVersion()
		latestAgentVersion := LatestAgentVersion()

		fmt.Println("Txlog Agent v" + agentVersion)
		fmt.Println("Txlog Server v" + serverVersion)

		if latestAgentVersion != "" && latestAgentVersion != "v"+agentVersion {
			fmt.Println("")
			fmt.Println("Your version of Txlog Agent is out of date! The latest version")
			fmt.Println("is " + latestAgentVersion + ". Go to https://txlog.rda.run/agent/latest for details.")
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// ServerVersion retrieves the server version from the configured server URL.
// It uses the resty library to make an HTTP GET request to the "/v1/version" endpoint.
// If authentication is configured (username and password), it sets the Basic Auth header.
// On success, it returns the server version string.
// On failure (including network errors or invalid server response), it returns "unknown".
func GetServerVersion() string {
	client := resty.New()
	client.SetAllowGetMethodPayload(true)

	var server ServerVersion
	req := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&server)

	if username := viper.GetString("server.username"); username != "" {
		req.SetBasicAuth(username, viper.GetString("server.password"))
	}

	_, err := req.Get(viper.GetString("server.url") + "/v1/version")

	if err == nil {
		return server.Version
	}

	return "unknown"
}

// LatestAgentVersion retrieves the latest agent version from a remote server.
// It checks the 'agent.check_version' configuration to determine if version checking is enabled.
// If enabled, it makes an HTTP GET request to "https://txlog.rda.run/docs/agent/version".
// If the request is successful and returns a 200 status code, the function returns the body of the response as a string,
// which represents the latest agent version.
// If there is an error during the request or the status code is not 200, or if version checking is disabled,
// the function returns an empty string.
func LatestAgentVersion() string {
	checkVersion := viper.Get("agent.check_version")

	if checkVersion == nil || checkVersion == true {
		response, err := resty.New().R().Get("https://txlog.rda.run/agent/version")

		if err == nil {
			if response.StatusCode() == 200 {
				return string(response.Body())
			}
		}
	}

	return ""
}
