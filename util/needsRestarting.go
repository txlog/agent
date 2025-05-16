package util

import (
	"bytes"
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
//
// The function parses the command output looking for the specific phrase
// "Reboot should not be necessary". If this phrase is found, it indicates
// no restart is needed.
func NeedsRestarting() (bool, string) {
	cmd := exec.Command(PackageBinary(), "needs-restarting", "-r")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	_ = cmd.Run()
	result := stdout.String() + stderr.String()

	if strings.Contains(result, "Reboot should not be necessary") {
		return false, result
	} else {
		return true, result
	}
}
