package cmd

import (
	"testing"

	"github.com/Masterminds/semver/v3"
)

// TestVersionComparison tests the semver logic used in ValidateServerVersionForAPIKey
func TestVersionComparison_Compatible(t *testing.T) {
	minVersion, _ := semver.NewVersion("1.14.0")

	compatibleVersions := []string{"1.14.0", "1.14.1", "1.15.0", "2.0.0", "10.0.0"}

	for _, versionStr := range compatibleVersions {
		version, err := semver.NewVersion(versionStr)
		if err != nil {
			t.Errorf("Failed to parse version %s: %v", versionStr, err)
			continue
		}

		if version.LessThan(minVersion) {
			t.Errorf("Version %s should be compatible (>= 1.14.0) but was marked as incompatible", versionStr)
		}
	}
}

func TestVersionComparison_Incompatible(t *testing.T) {
	minVersion, _ := semver.NewVersion("1.14.0")

	incompatibleVersions := []string{"1.13.9", "1.13.0", "1.0.0", "0.9.0"}

	for _, versionStr := range incompatibleVersions {
		version, err := semver.NewVersion(versionStr)
		if err != nil {
			t.Errorf("Failed to parse version %s: %v", versionStr, err)
			continue
		}

		if !version.LessThan(minVersion) {
			t.Errorf("Version %s should be incompatible (< 1.14.0) but was marked as compatible", versionStr)
		}
	}
}

func TestVersionComparison_InvalidFormats(t *testing.T) {
	invalidVersions := []string{"invalid", "1.x.0", "abc", ""}

	for _, versionStr := range invalidVersions {
		_, err := semver.NewVersion(versionStr)
		if err == nil {
			t.Errorf("Version %s should be invalid but was parsed successfully", versionStr)
		}
	}
}

func TestVersionComparison_EdgeCases(t *testing.T) {
	minVersion, _ := semver.NewVersion("1.14.0")

	// Test boundary case - exactly at minimum
	version, _ := semver.NewVersion("1.14.0")
	if version.LessThan(minVersion) {
		t.Error("Version 1.14.0 should be considered compatible (equal to minimum)")
	}

	// Test just below minimum
	version, _ = semver.NewVersion("1.13.99")
	if !version.LessThan(minVersion) {
		t.Error("Version 1.13.99 should be considered incompatible (less than minimum)")
	}
}
