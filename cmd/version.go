package cmd

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/txlog/agent/util"
)

const agentVersion = "1.8.0"

type ServerVersion struct {
	Version string `json:"version"`
}

// ServerVersionError represents an error when fetching server version
type ServerVersionError struct {
	StatusCode int
	Message    string
}

func (e *ServerVersionError) Error() string {
	return e.Message
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

// GetServerVersionWithError retrieves the server version and returns detailed error information.
// Returns the version string and any error encountered.
// If successful, returns version and nil error.
// If there's an authentication error, returns empty string and ServerVersionError with status code.
// If there's a network error, returns empty string and the network error.
func GetServerVersionWithError() (string, error) {
	client := resty.New()
	client.SetAllowGetMethodPayload(true)

	var server ServerVersion
	req := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&server)

	util.SetAuthentication(req)

	resp, err := req.Get(viper.GetString("server.url") + "/v1/version")

	// Network error
	if err != nil {
		return "", fmt.Errorf("failed to connect to server: %w", err)
	}

	// Authentication or other HTTP errors
	if resp.StatusCode() != 200 {
		if resp.StatusCode() == 401 {
			return "", &ServerVersionError{
				StatusCode: 401,
				Message:    "authentication failed: invalid credentials (API key or username/password)",
			}
		}
		// Truncate response body to first 100 characters to avoid leaking sensitive details
		safeBody := resp.String()
		if len(safeBody) > 100 {
			safeBody = safeBody[:100] + "..."
		}
		return "", &ServerVersionError{
			StatusCode: resp.StatusCode(),
			Message:    fmt.Sprintf("server returned status %d: %s", resp.StatusCode(), safeBody),
		}
	}

	return server.Version, nil
}

// GetServerVersion retrieves the server version from the configured server URL.
// It uses the resty library to make an HTTP GET request to the "/v1/version" endpoint.
// If authentication is configured (API key or username and password), it sets the appropriate headers.
// On success, it returns the server version string.
// On failure (including network errors or invalid server response), it returns "unknown".
func GetServerVersion() string {
	version, err := GetServerVersionWithError()
	if err != nil {
		return "unknown"
	}
	return version
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

// ValidateServerVersionForAPIKey checks if the server version supports API key authentication.
// API key authentication requires server version >= 1.14.0.
// Returns an error if the server version is too old or cannot be determined.
func ValidateServerVersionForAPIKey() error {
	serverVersion, err := GetServerVersionWithError()

	// If there's an error getting the version, return it directly
	if err != nil {
		// Check if it's an authentication error
		if serverErr, ok := err.(*ServerVersionError); ok {
			return serverErr
		}
		return err
	}

	// If version is empty (shouldn't happen with above checks, but just in case)
	if serverVersion == "" {
		return fmt.Errorf("server returned empty version. API key authentication requires server version >= 1.14.0")
	}

	// Parse server version
	sv, err := semver.NewVersion(serverVersion)
	if err != nil {
		return fmt.Errorf("invalid server version format '%s'. API key authentication requires server version >= 1.14.0", serverVersion)
	}

	// Minimum version required for API key support
	minVersion := semver.MustParse("1.14.0")

	// Check if server version is compatible
	if sv.LessThan(minVersion) {
		return fmt.Errorf("server version %s does not support API key authentication. Please upgrade the server to version 1.14.0 or higher, or use basic authentication instead", serverVersion)
	}

	return nil
}
