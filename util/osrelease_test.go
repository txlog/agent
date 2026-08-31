package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOSRelease(t *testing.T) {
	content := `# this is a comment
NAME="Rocky Linux"

VERSION_ID="9.6"
PRETTY_NAME="Rocky Linux 9.6 (Blue Onyx)"
ID=rocky
ANSI_COLOR='0;32'
CPE_NAME="cpe:/o:rocky:rocky:9::baseos"
malformed line without separator
`
	path := filepath.Join(t.TempDir(), "os-release")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	original := Path
	Path = path
	defer func() { Path = original }()

	release, err := OSRelease()
	if err != nil {
		t.Fatalf("OSRelease() returned %v", err)
	}

	want := map[string]string{
		"NAME":        "Rocky Linux",
		"VERSION_ID":  "9.6",
		"PRETTY_NAME": "Rocky Linux 9.6 (Blue Onyx)",
		"ID":          "rocky",
		"ANSI_COLOR":  "0;32",
		"CPE_NAME":    "cpe:/o:rocky:rocky:9::baseos", // value containing colons and no quotes issue
	}
	for key, expected := range want {
		if got := release[key]; got != expected {
			t.Errorf("%s = %q, want %q", key, got, expected)
		}
	}

	if _, found := release["malformed line without separator"]; found {
		t.Error("line without '=' should have been skipped")
	}
}

func TestOSReleaseMissingFile(t *testing.T) {
	original := Path
	Path = filepath.Join(t.TempDir(), "does-not-exist")
	defer func() { Path = original }()

	if _, err := OSRelease(); err == nil {
		t.Error("expected an error for a missing os-release file")
	}
}
