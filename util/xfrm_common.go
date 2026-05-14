package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

// XFRMModuleStatus holds the result of checking XFRM/ESP kernel modules.
type XFRMModuleStatus struct {
	ESP4Loaded  bool
	ESP6Loaded  bool
	RxRPCLoaded bool
	AnyLoaded   bool
}

// CheckXFRMModules checks if the esp4, esp6 and rxrpc kernel modules are loaded.
// It reads /proc/modules and also checks /sys/module/ for each module.
func CheckXFRMModules() (XFRMModuleStatus, []string) {
	var status XFRMModuleStatus
	var details []string

	modules := map[string]*bool{
		"esp4":  &status.ESP4Loaded,
		"esp6":  &status.ESP6Loaded,
		"rxrpc": &status.RxRPCLoaded,
	}

	// Check /proc/modules first
	f, err := os.Open("/proc/modules")
	if err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) > 0 {
				if ptr, ok := modules[fields[0]]; ok {
					*ptr = true
				}
			}
		}
	}

	// Also check /sys/module/ as fallback (module may be built-in)
	for name, ptr := range modules {
		if !*ptr {
			if _, err := os.Stat(fmt.Sprintf("/sys/module/%s", name)); err == nil {
				*ptr = true
			}
		}
	}

	status.AnyLoaded = status.ESP4Loaded || status.ESP6Loaded || status.RxRPCLoaded

	if status.ESP4Loaded {
		details = append(details, "Module esp4 is loaded")
	}
	if status.ESP6Loaded {
		details = append(details, "Module esp6 is loaded")
	}
	if status.RxRPCLoaded {
		details = append(details, "Module rxrpc is loaded")
	}
	if !status.AnyLoaded {
		details = append(details, "No XFRM/ESP modules loaded (esp4, esp6, rxrpc)")
	}

	return status, details
}

// CheckDirtyFragMitigation checks if the dirtyfrag mitigation file exists
// at /etc/modprobe.d/dirtyfrag.conf and contains the expected content.
func CheckDirtyFragMitigation() (bool, []string) {
	var details []string
	mitigationPath := "/etc/modprobe.d/dirtyfrag.conf"

	data, err := os.ReadFile(mitigationPath)
	if err != nil {
		details = append(details, fmt.Sprintf("No mitigation file found at %s", mitigationPath))
		return false, details
	}

	content := string(data)
	if strings.Contains(content, "install esp4 /bin/false") {
		details = append(details, fmt.Sprintf("Mitigation active: %s contains esp4 block", mitigationPath))
		return true, details
	}

	details = append(details, fmt.Sprintf("Mitigation file exists at %s but does not block esp4", mitigationPath))
	return false, details
}

// GetKernelVersion returns the running kernel version string from uname.
func GetKernelVersion() string {
	var utsname unix.Utsname
	if err := unix.Uname(&utsname); err != nil {
		return ""
	}

	// Convert [65]byte to string, trimming null bytes
	release := make([]byte, 0, len(utsname.Release))
	for _, b := range utsname.Release {
		if b == 0 {
			break
		}
		release = append(release, byte(b))
	}
	return string(release)
}
