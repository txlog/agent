package util

// The original implementation of this parser was derived from the osrelease
// package by Dániel Görbe (MIT licensed, 2021). Only the file format it reads
// remains: https://www.linux.org/docs/man5/os-release.html

import (
	"fmt"
	"os"
	"strings"
)

// Path contains the default path to the os-release file
var Path = "/etc/os-release"

// OSRelease parses the os-release file pointed to by Path and returns its
// fields, with surrounding quotes stripped from the values. Empty lines and
// comments are ignored, as are lines without a '='.
func OSRelease() (map[string]string, error) {
	data, err := os.ReadFile(Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", Path, err)
	}

	fields := make(map[string]string)
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if key, value, found := strings.Cut(line, "="); found {
			fields[key] = strings.Trim(value, `"'`)
		}
	}

	return fields, nil
}
