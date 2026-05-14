package util

import (
	"fmt"
	"strings"
)

// DirtyFragResult represents the result of Dirty Frag vulnerability detection.
type DirtyFragResult struct {
	Vulnerable  bool   // true if system is likely vulnerable
	Description string // human-readable summary
}

// CheckDirtyFrag performs a non-destructive detection of the Dirty Frag vulnerability.
//
// Dirty Frag exploits a bug in the Linux XFRM ESP-in-UDP subsystem that allows
// writing arbitrary data into the kernel page cache of read-only files via the
// seq_hi field of ESP Extended Sequence Number (ESN) processing.
//
// Detection is based on pre-conditions:
//  1. XFRM/ESP modules (esp4, esp6, rxrpc) must be loaded or loadable
//  2. Kernel must not contain the fix commit f4c50a4034e6
//  3. Mitigation file /etc/modprobe.d/dirtyfrag.conf must not be present
func CheckDirtyFrag() DirtyFragResult {
	// Check if XFRM/ESP modules are available
	moduleStatus, _ := CheckXFRMModules()
	if !moduleStatus.AnyLoaded {
		return DirtyFragResult{
			Vulnerable:  false,
			Description: "Not vulnerable: no XFRM/ESP modules loaded",
		}
	}

	// Check if mitigation is applied
	mitigated, _ := CheckDirtyFragMitigation()
	if mitigated {
		return DirtyFragResult{
			Vulnerable:  false,
			Description: "Not vulnerable: mitigation applied (dirtyfrag.conf)",
		}
	}

	// Check kernel version for the fix
	kernelVersion := GetKernelVersion()
	if kernelVersion == "" {
		return DirtyFragResult{
			Vulnerable:  false,
			Description: "Inconclusive: unable to determine kernel version",
		}
	}

	if isDirtyFragPatched(kernelVersion) {
		return DirtyFragResult{
			Vulnerable:  false,
			Description: fmt.Sprintf("Not vulnerable: kernel %s contains the fix", kernelVersion),
		}
	}

	return DirtyFragResult{
		Vulnerable:  true,
		Description: fmt.Sprintf("Vulnerable: kernel %s lacks Dirty Frag fix, XFRM modules loaded, no mitigation", kernelVersion),
	}
}

// isDirtyFragPatched checks if the kernel version includes the Dirty Frag fix.
// The fix commit is f4c50a4034e62ab75f1d5cdd191dd5f9c77fdff4 ("xfrm: esp:
// avoid in-place decrypt on shared skb frags"). This was merged into net.git
// on 2026-05-07. Stable backports follow the standard kernel release cadence.
//
// Since there's no reliable way to check for a specific commit in a running
// kernel, we use a conservative heuristic: kernels released before 2026-05-07
// are assumed unpatched unless they are known-fixed stable releases.
// For RHEL/CentOS/Alma/Rocky, the fix will come via errata kernel updates.
func isDirtyFragPatched(version string) bool {
	// Check for known-patched upstream stable versions
	// These will be updated as distros release fixes
	patchedPrefixes := []string{
		// Placeholder for future patched versions from distros
		// e.g., "5.15.160", "6.1.95", "6.6.35", "6.9.6"
	}

	for _, prefix := range patchedPrefixes {
		if strings.HasPrefix(version, prefix) {
			return true
		}
	}

	return false
}
