package util

import (
	"fmt"
	"strings"
)

// FragnesiaResult represents the result of Fragnesia vulnerability detection.
type FragnesiaResult struct {
	Vulnerable  bool   // true if system is likely vulnerable
	Description string // human-readable summary
}

// CheckFragnesia performs a non-destructive detection of the Fragnesia vulnerability.
//
// Fragnesia exploits a logic bug in the Linux XFRM ESP-in-TCP subsystem where
// skb_try_coalesce() loses the SKBFL_SHARED_FRAG marker when transferring paged
// frags. This allows ESP to decrypt in-place over page-cache-backed frags,
// enabling arbitrary byte writes via AES-GCM keystream XOR.
//
// The fix is the patch "net: skbuff: preserve shared-frag marker during
// coalescing" submitted 2026-05-13 to netdev.
//
// Detection is based on pre-conditions:
//  1. XFRM/ESP modules (esp4, esp6, rxrpc) must be loaded or loadable
//  2. Kernel must not contain the coalescing fix
//  3. Mitigation file /etc/modprobe.d/dirtyfrag.conf must not be present
//     (same mitigation as Dirty Frag — rmmod esp4 esp6 rxrpc)
func CheckFragnesia() FragnesiaResult {
	// Check if XFRM/ESP modules are available
	moduleStatus, _ := CheckXFRMModules()
	if !moduleStatus.AnyLoaded {
		return FragnesiaResult{
			Vulnerable:  false,
			Description: "Not vulnerable: no XFRM/ESP modules loaded",
		}
	}

	// Check if mitigation is applied (same as Dirty Frag)
	mitigated, _ := CheckDirtyFragMitigation()
	if mitigated {
		return FragnesiaResult{
			Vulnerable:  false,
			Description: "Not vulnerable: mitigation applied (dirtyfrag.conf)",
		}
	}

	// Check kernel version for the fix
	kernelVersion := GetKernelVersion()
	if kernelVersion == "" {
		return FragnesiaResult{
			Vulnerable:  false,
			Description: "Inconclusive: unable to determine kernel version",
		}
	}

	if isFragnesiaPatched(kernelVersion) {
		return FragnesiaResult{
			Vulnerable:  false,
			Description: fmt.Sprintf("Not vulnerable: kernel %s contains the fix", kernelVersion),
		}
	}

	return FragnesiaResult{
		Vulnerable:  true,
		Description: fmt.Sprintf("Vulnerable: kernel %s lacks Fragnesia fix, XFRM modules loaded, no mitigation", kernelVersion),
	}
}

// isFragnesiaPatched checks if the kernel version includes the Fragnesia fix.
// The fix is the patch "net: skbuff: preserve shared-frag marker during
// coalescing" which propagates SKBFL_SHARED_FRAG in skb_try_coalesce().
// Submitted to netdev on 2026-05-13. This is a SEPARATE fix from the
// Dirty Frag commit — a kernel may have one fix but not the other.
func isFragnesiaPatched(version string) bool {
	// Check for known-patched upstream stable versions
	// These will be updated as distros release fixes
	patchedPrefixes := []string{
		// Placeholder for future patched versions from distros
		// The fragnesia patch was submitted 2026-05-13, so no stable
		// releases contain it yet.
	}

	for _, prefix := range patchedPrefixes {
		if strings.HasPrefix(version, prefix) {
			return true
		}
	}

	return false
}
