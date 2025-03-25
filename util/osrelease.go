// Package osrelease parse /etc/os-release
// MIT License
//
// Copyright (c) 2021 Dániel Görbe

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
// More about the os-release: https://www.linux.org/docs/man5/os-release.html
package util

import (
	"fmt"
	"os"
	"strings"
)

// Path contains the default path to the os-release file
var Path = "/etc/os-release"

var Release OSRelease

type OSRelease struct {
	Name             string
	Version          string
	ID               string
	IDLike           string
	PrettyName       string
	VersionID        string
	HomeURL          string
	DocumentationURL string
	SupportURL       string
	BugReportURL     string
	PrivacyPolicyURL string
	VersionCodename  string
	UbuntuCodename   string
	ANSIColor        string
	CPEName          string
	BuildID          string
	Variant          string
	VariantID        string
	Logo             string
}

// getLines read the OSReleasePath and return it line by line.
// Empty lines and comments (beginning with a "#") are ignored.
func getLines() ([]string, error) {

	output, err := os.ReadFile(Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %s", Path, err)
	}

	lines := make([]string, 0)

	for _, line := range strings.Split(string(output), "\n") {

		switch true {
		case line == "":
			continue
		case []byte(line)[0] == '#':
			continue
		}

		lines = append(lines, line)
	}

	return lines, nil
}

// parseLine parse a single line.
// Return key, value, error (if any)
func parseLine(line string) (string, string, error) {

	subs := strings.SplitN(line, "=", 2)

	if len(subs) != 2 {
		return "", "", fmt.Errorf("invalid length of the substrings: %d", len(subs))
	}

	return subs[0], strings.Trim(subs[1], "\"'"), nil
}

// ParseOSRelease parses the os-release file pointing to by Path.
// The fields are saved into the Release global variable.
func ParseOSRelease() error {

	lines, err := getLines()
	if err != nil {
		return fmt.Errorf("failed to get lines of %s: %s", Path, err)
	}

	for i := range lines {

		key, value, err := parseLine(lines[i])
		if err != nil {
			return fmt.Errorf("failed to parse line '%s': %s", lines[i], err)
		}

		switch key {
		case "NAME":
			Release.Name = value
		case "VERSION":
			Release.Version = value
		case "ID":
			Release.ID = value
		case "ID_LIKE":
			Release.IDLike = value
		case "PRETTY_NAME":
			Release.PrettyName = value
		case "VERSION_ID":
			Release.VersionID = value
		case "HOME_URL":
			Release.HomeURL = value
		case "DOCUMENTATION_URL":
			Release.DocumentationURL = value
		case "SUPPORT_URL":
			Release.SupportURL = value
		case "BUG_REPORT_URL":
			Release.BugReportURL = value
		case "PRIVACY_POLICY_URL":
			Release.PrivacyPolicyURL = value
		case "VERSION_CODENAME":
			Release.VersionCodename = value
		case "UBUNTU_CODENAME":
			Release.UbuntuCodename = value
		case "ANSI_COLOR":
			Release.ANSIColor = value
		case "CPE_NAME":
			Release.CPEName = value
		case "BUILD_ID":
			Release.BuildID = value
		case "VARIANT":
			Release.Variant = value
		case "VARIANT_ID":
			Release.VariantID = value
		case "LOGO":
			Release.Logo = value
		}
	}

	return nil
}
