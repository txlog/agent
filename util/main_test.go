package util

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
)

func TestSetAuthentication_WithAPIKey(t *testing.T) {
	// Setup
	viper.Reset()
	viper.Set("server.api_key", "txlog_test_key_123")
	viper.Set("server.username", "testuser")
	viper.Set("server.password", "testpass")

	client := resty.New()
	req := client.R()

	// Execute
	SetAuthentication(req)

	// Verify API key header is set (preferred method)
	if req.Header.Get("X-API-Key") != "txlog_test_key_123" {
		t.Errorf("Expected X-API-Key header to be 'txlog_test_key_123', got '%s'", req.Header.Get("X-API-Key"))
	}

	// Verify basic auth is not set when API key is present
	if req.Header.Get("Authorization") != "" {
		t.Errorf("Expected Authorization header to be empty when API key is set, got '%s'", req.Header.Get("Authorization"))
	}
}

func TestSetAuthentication_WithBasicAuth(t *testing.T) {
	// Setup
	viper.Reset()
	viper.Set("server.username", "testuser")
	viper.Set("server.password", "testpass")

	client := resty.New()
	req := client.R()

	// Execute
	SetAuthentication(req)

	// Verify API key header is not set
	if req.Header.Get("X-API-Key") != "" {
		t.Errorf("Expected X-API-Key header to be empty when using basic auth, got '%s'", req.Header.Get("X-API-Key"))
	}

	// Note: We can't directly verify the Authorization header here because resty's SetBasicAuth
	// sets it internally and only applies it when the request is actually sent.
	// The important thing is that we verify API key is not set, which means basic auth was configured.
}

func TestSetAuthentication_WithNoAuth(t *testing.T) {
	// Setup
	viper.Reset()

	client := resty.New()
	req := client.R()

	// Execute
	SetAuthentication(req)

	// Verify no auth headers are set
	if req.Header.Get("X-API-Key") != "" {
		t.Errorf("Expected X-API-Key header to be empty, got '%s'", req.Header.Get("X-API-Key"))
	}
	if req.Header.Get("Authorization") != "" {
		t.Errorf("Expected Authorization header to be empty, got '%s'", req.Header.Get("Authorization"))
	}
}

func TestSetAuthentication_APIKeyPreferredOverBasicAuth(t *testing.T) {
	// Setup - both API key and basic auth configured
	viper.Reset()
	viper.Set("server.api_key", "txlog_preferred_key")
	viper.Set("server.username", "testuser")
	viper.Set("server.password", "testpass")

	client := resty.New()
	req := client.R()

	// Execute
	SetAuthentication(req)

	// Verify API key is used (not basic auth)
	if req.Header.Get("X-API-Key") != "txlog_preferred_key" {
		t.Errorf("Expected X-API-Key header to be 'txlog_preferred_key', got '%s'", req.Header.Get("X-API-Key"))
	}

	// When API key is set, basic auth should not be used
	// (We can't verify the Authorization header directly, but we know the function
	// returns early when API key is set, so basic auth is never configured)
}

// TestSplitPackageName covers the NEVRA shapes dnf produces, including the
// truncated ones that hit the function's early returns.
func TestSplitPackageName(t *testing.T) {
	tests := []struct {
		packageName                            string
		name, version, release, epoch, archStr string
	}{
		{
			packageName: "bash-5.1.8-9.el9.x86_64",
			name:        "bash", version: "5.1.8", release: "9.el9", epoch: "", archStr: "x86_64",
		},
		{
			// The epoch belongs in its own field, not glued to the version.
			packageName: "grub2-common-1:2.06-95.el9.noarch",
			name:        "grub2-common", version: "2.06", release: "95.el9", epoch: "1", archStr: "noarch",
		},
		{
			// Name with a dash, release with several dots.
			packageName: "kernel-core-5.14.0-570.12.1.el9_6.x86_64",
			name:        "kernel-core", version: "5.14.0", release: "570.12.1.el9_6", epoch: "", archStr: "x86_64",
		},
		{
			// The .rpm suffix is optional and trimmed.
			packageName: "python3-dnf-4.14.0-9.el9.noarch.rpm",
			name:        "python3-dnf", version: "4.14.0", release: "9.el9", epoch: "", archStr: "noarch",
		},
		{
			// No dot at all: everything is the name.
			packageName: "some-package",
			name:        "some-package", version: "", release: "", epoch: "", archStr: "",
		},
		{
			// A dot but no dash: name plus arch only.
			packageName: "package.x86_64",
			name:        "package", version: "", release: "", epoch: "", archStr: "x86_64",
		},
		{
			// One dash: no version to be found, only a release.
			packageName: "package-1.el9.noarch",
			name:        "package", version: "", release: "1.el9", epoch: "", archStr: "noarch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.packageName, func(t *testing.T) {
			name, version, release, epoch, arch := SplitPackageName(tt.packageName)
			for _, f := range []struct{ field, got, want string }{
				{"name", name, tt.name},
				{"version", version, tt.version},
				{"release", release, tt.release},
				{"epoch", epoch, tt.epoch},
				{"arch", arch, tt.archStr},
			} {
				if f.got != f.want {
					t.Errorf("%s = %q, want %q", f.field, f.got, f.want)
				}
			}
		})
	}
}
