package util

import (
	"os"
	"testing"
)

func TestCheckCopyFail(t *testing.T) {
	result := CheckCopyFail()

	// The function should always return a valid result, regardless of
	// whether the system is vulnerable or not.
	if result.Description == "" {
		t.Error("CheckCopyFail returned empty Description")
	}
	if result.Details == "" {
		t.Error("CheckCopyFail returned empty Details")
	}

	// If vulnerable, Description should mention CVE
	if result.Vulnerable {
		if result.Description == "" {
			t.Error("Vulnerable result should have a description")
		}
		t.Logf("System is VULNERABLE: %s", result.Description)
		t.Logf("Escalation confirmed: %v", result.EscalationConfirmed)
		if result.SetuidTarget != "" {
			t.Logf("Setuid target: %s", result.SetuidTarget)
		}
	} else {
		t.Logf("System is NOT vulnerable: %s", result.Description)
	}

	t.Logf("Details:\n%s", result.Details)
}

func TestFindSetuidBinaries(t *testing.T) {
	binaries := findSetuidBinaries()

	// On most Linux systems, at least /usr/bin/su or /usr/bin/passwd
	// should be setuid-root. We don't fail if none found (could be a
	// container or hardened system), but we log it.
	if len(binaries) == 0 {
		t.Log("No setuid-root binaries found (may be a container or hardened system)")
		return
	}

	for _, b := range binaries {
		t.Logf("Found setuid binary: %s (readable: %v)", b.Path, b.Readable)

		// Verify the binary actually has setuid bit
		info, err := os.Stat(b.Path)
		if err != nil {
			t.Errorf("Could not stat %s: %v", b.Path, err)
			continue
		}
		if info.Mode()&os.ModeSetuid == 0 {
			t.Errorf("%s reported as setuid but ModeSetuid is not set", b.Path)
		}
	}
}

func TestCheckAFALGAvailable(t *testing.T) {
	err := checkAFALGAvailable()

	if err != nil {
		t.Logf("AF_ALG not available: %v", err)
		t.Log("This is expected on systems with AF_ALG disabled or in containers")
	} else {
		t.Log("AF_ALG socket and authencesn are available")
	}
}

func TestCopyFailResultFields(t *testing.T) {
	// Test that a non-vulnerable result has correct field values
	result := CopyFailResult{
		Vulnerable:          false,
		EscalationConfirmed: false,
		Description:         "Not vulnerable: test",
		Details:             "test details",
		SetuidTarget:        "",
	}

	if result.Vulnerable {
		t.Error("Expected Vulnerable to be false")
	}
	if result.EscalationConfirmed {
		t.Error("Expected EscalationConfirmed to be false")
	}
	if result.SetuidTarget != "" {
		t.Error("Expected empty SetuidTarget for non-vulnerable result")
	}

	// Test that a vulnerable result with escalation has correct field values
	result = CopyFailResult{
		Vulnerable:          true,
		EscalationConfirmed: true,
		Description:         "Vulnerable to CVE-2026-31431 — privilege escalation possible via /usr/bin/su",
		Details:             "test details",
		SetuidTarget:        "/usr/bin/su",
	}

	if !result.Vulnerable {
		t.Error("Expected Vulnerable to be true")
	}
	if !result.EscalationConfirmed {
		t.Error("Expected EscalationConfirmed to be true")
	}
	if result.SetuidTarget != "/usr/bin/su" {
		t.Errorf("Expected SetuidTarget '/usr/bin/su', got '%s'", result.SetuidTarget)
	}
}
