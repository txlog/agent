package util

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/go-resty/resty/v2"
	"github.com/itlightning/dateparse"
	"github.com/spf13/viper"
)

// DateConversion takes a date string input and converts it to RFC3339 format.
// It attempts to parse the input date string using multiple common date formats
// and returns the formatted date as a string.
//
// Parameters:
//   - data: A string representing a date in any common format
//
// Returns:
//   - string: The date formatted in RFC3339 format (2006-01-02T15:04:05Z07:00)
//   - error: An error if the date parsing fails, nil otherwise
func DateConversion(data string) (string, error) {
	t, err := dateparse.ParseLocal(data)
	if err != nil {
		return "", err
	}

	formattedDate := t.Format("2006-01-02T15:04:05Z07:00")

	return formattedDate, nil
}

// GetMachineId retrieves the unique machine identifier from the '/etc/machine-id' file.
// This identifier is typically used to distinguish the host machine in a network.
//
// Returns:
//   - string: The machine ID as a trimmed string
//   - error: An error if reading the machine-id file fails
func GetMachineId() (string, error) {
	data, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		return "", fmt.Errorf("error while reading /etc/machine-id: %w", err)
	}

	machineID := strings.TrimSpace(string(data))

	return machineID, nil
}

// GetHostname returns the system hostname by executing the hostname command.
// It uses the /usr/bin/hostname command to obtain the hostname and returns it as a string.
// Returns an error if the hostname command execution fails.
func GetHostname() (string, error) {
	out, err := exec.Command("/usr/bin/hostname").Output()
	if err != nil {
		return "", fmt.Errorf("error while executing hostname command: %w", err)
	}

	hostname := strings.TrimSpace(string(out))

	return hostname, nil
}

// SplitPackageName splits a RPM package name into its components.
// It takes a package name string as input and returns the following components:
//   - name: The name of the package
//   - version: The version number
//   - release: The release number
//   - epoch: The epoch number (empty string if not present)
//   - arch: The architecture
//
// The function expects package names in the following format:
// [name]-[version]-[release].[arch].rpm
// or
// [name]-[epoch]:[version]-[release].[arch].rpm
//
// The .rpm suffix is optional and will be trimmed if present.
// If epoch is not present in the package name, an empty string is returned for that component.
func SplitPackageName(package_name string) (name, version, release, epoch, arch string) {
	package_name = strings.TrimSuffix(package_name, ".rpm")

	archIndex := strings.LastIndex(package_name, ".")
	arch = package_name[archIndex+1:]

	relIndex := strings.LastIndex(package_name[:archIndex], "-")
	release = package_name[relIndex+1 : archIndex]

	verIndex := strings.LastIndex(package_name[:relIndex], "-")
	version = package_name[verIndex+1 : relIndex]

	epochIndex := strings.Index(package_name, ":")
	if epochIndex == -1 {
		epoch = ""
		name = package_name[0:verIndex]
	} else {
		epoch = package_name[strings.LastIndex(package_name[:relIndex], "-")+1 : epochIndex]
		name = package_name[:verIndex]
	}

	return
}

// PackageBinary determines and verifies the appropriate package manager binary (yum or dnf)
// based on the Linux distribution version. It reads /etc/os-release to check if the system
// is running RHEL/CentOS 8 or 9, in which case it selects 'dnf' instead of the default 'yum'.
//
// The function also verifies if the selected package manager is installed in the system.
// If the binary is not found, it exits with an error message.
//
// Returns:
//   - string: The name of the package manager binary ("yum" or "dnf")
//
// The function will exit with status code 1 if the required package manager is not installed.
func PackageBinary() string {
	binary := "yum"

	// Read /etc/os-release line-by-line
	releaseData, _ := os.ReadFile("/etc/os-release")
	for _, line := range strings.Split(string(releaseData), "\n") {
		re := regexp.MustCompile(`VERSION_ID="([1-9][0-9]?)(?:\.[0-9]+)?"`)
		match := re.FindStringSubmatch(line)
		if len(match) > 1 {
			versionID := match[1]
			vID, _ := strconv.Atoi(versionID)
			if vID >= 8 {
				binary = "dnf"
				break
			}
		}
	}

	if !binaryInstalled(binary) {
		color.Red("ERROR: %s is not installed. Exiting.", binary)
		os.Exit(1)
	}

	return binary
}

// binaryInstalled checks if a binary is installed in the system.
// It takes a binary name as input and returns true if the binary is found in the system PATH,
// false otherwise.
//
// Parameters:
//   - binaryName: string representing the name of the binary to check
//
// Returns:
//   - bool: true if binary is installed, false if not found
func binaryInstalled(binaryName string) bool {
	_, err := exec.LookPath(binaryName)
	return err == nil
}

// SetAuthentication configures authentication for an API request.
// It prioritizes API key authentication over basic authentication.
// If an API key is configured in the server.api_key setting, it sets the X-API-Key header.
// Otherwise, if username and password are configured, it uses basic authentication.
//
// Parameters:
//   - request: A resty.Request instance to configure with authentication headers
func SetAuthentication(request *resty.Request) {
	// Check for API key first (preferred method)
	if apiKey := viper.GetString("server.api_key"); apiKey != "" {
		request.SetHeader("X-API-Key", apiKey)
		return
	}

	// Fall back to basic authentication if configured
	if username := viper.GetString("server.username"); username != "" {
		request.SetBasicAuth(username, viper.GetString("server.password"))
	}
}
