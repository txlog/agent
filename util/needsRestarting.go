package util

import (
	"os/exec"
	"strings"
)

// NeedsRestarting checks if the system needs to be restarted by executing the
// '/usr/bin/needs-restarting' command. This command is typically available on
// Red Hat-based systems to determine if any running processes are using files
// that have been updated/deleted.
//
// Returns:
//   - bool: true if system needs restarting, false otherwise
//   - string: the complete output message from needs-restarting command
//   - error: any error encountered while executing the command
//
// The function parses the command output looking for the specific phrase
// "Reboot should not be necessary". If this phrase is found, it indicates
// no restart is needed.
func NeedsRestarting() (bool, string, error) {
	out, err := exec.Command(PackageBinary(), "needs-restarting", "-r").Output()
	if err != nil {
		return false, "", err
	}

	result := strings.TrimSpace(string(out))

	if strings.Contains(result, "Reboot should not be necessary") {
		return false, result, nil
	} else {
		return true, result, nil
	}
}
