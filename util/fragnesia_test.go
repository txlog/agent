package util

import (
	"testing"
)

func TestCheckFragnesia(t *testing.T) {
	result := CheckFragnesia()

	// The function should always return a valid result
	if result.Description == "" {
		t.Error("CheckFragnesia returned empty Description")
	}

	if result.Vulnerable {
		t.Logf("System is VULNERABLE: %s", result.Description)
	} else {
		t.Logf("System is NOT vulnerable: %s", result.Description)
	}
}

func TestFragnesiaResultFields(t *testing.T) {
	result := FragnesiaResult{
		Vulnerable:  false,
		Description: "Not vulnerable: test",
	}

	if result.Vulnerable {
		t.Error("Expected Vulnerable to be false")
	}

	result = FragnesiaResult{
		Vulnerable:  true,
		Description: "Vulnerable: test",
	}

	if !result.Vulnerable {
		t.Error("Expected Vulnerable to be true")
	}
}

func TestIsFragnesiaPatched(t *testing.T) {
	// Currently no patched versions are registered,
	// so all versions should return false
	tests := []struct {
		version string
		want    bool
	}{
		{"6.8.0-111-generic", false},
		{"5.14.0-362.el9.x86_64", false},
		{"6.1.90-1.el9.x86_64", false},
	}

	for _, tt := range tests {
		got := isFragnesiaPatched(tt.version)
		if got != tt.want {
			t.Errorf("isFragnesiaPatched(%q) = %v, want %v", tt.version, got, tt.want)
		}
	}
}
