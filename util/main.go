package util

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
)

func DateConversion(data string) (string, error) {
	t, err := time.Parse("Mon Jan 2 15:04:05 2006", data)
	if err != nil {
		return "", err
	}

	dataFormatada := t.Format("2006-01-02 15:04:05")

	return dataFormatada, nil
}

// GetMachineId reads the contents of /etc/machine-id file
func GetMachineId() (string, error) {
	data, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		return "", fmt.Errorf("error while reading /etc/machine-id: %w", err)
	}

	machineID := strings.TrimSpace(string(data))

	return machineID, nil
}

// GetHostname reads the contents of /etc/hostname file
func GetHostname() (string, error) {
	data, err := os.ReadFile("/etc/hostname")
	if err != nil {
		return "", fmt.Errorf("error while reading /etc/hostname: %w", err)
	}

	machineID := strings.TrimSpace(string(data))

	return machineID, nil
}

// SplitPackageName splits the given package name into its components.
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
	} else {
		epoch = package_name[:epochIndex]
	}

	name = package_name[epochIndex+1 : verIndex]
	return
}

// PackageBinary determines the package manager (dnf or yum) to use
func PackageBinary() string {
	binary := "yum"

	// Read /etc/os-release line-by-line
	releaseData, _ := os.ReadFile("/etc/os-release")
	for _, line := range strings.Split(string(releaseData), "\n") {
		if strings.Contains(line, "VERSION_ID=\"8") || strings.Contains(line, "VERSION_ID=\"9") {
			binary = "dnf"
			break
		}
	}

	if !binaryInstalled(binary) {
		color.Red("ERROR: %s is not installed. Exiting.", binary)
		os.Exit(1)
	}

	return binary
}

// binaryInstalled checks if a binary is installed using the 'which' command
func binaryInstalled(binaryName string) bool {
	out, _ := exec.Command("which", binaryName).CombinedOutput()
	return !strings.Contains(string(out), "no "+binaryName+" in")
}
