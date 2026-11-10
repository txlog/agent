package cmd

import "testing"

// TestParseHistoryLine feeds reHistoryLine the real output of
// 'dnf history list', including the two cases that used to break it: a
// transaction with a two-digit altered count and one with an empty command
// line.
func TestParseHistoryLine(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		id          string
		commandLine string
		actions     string
		altered     string
	}{
		{
			name:        "single digit altered",
			line:        "    82 | install gnome-remote-des | 2026-08-30 19:04 | Install        |    1   ",
			id:          "82",
			commandLine: "install gnome-remote-des",
			actions:     "Install",
			altered:     "1",
		},
		{
			name:        "two digit altered is not truncated",
			line:        "    80 | -y up                    | 2026-08-30 09:54 | Upgrade        |   22   ",
			id:          "80",
			commandLine: "-y up",
			actions:     "Upgrade",
			altered:     "22",
		},
		{
			name:        "empty command line",
			line:        "    81 |                          | 2026-08-30 19:03 | Install        |    4   ",
			id:          "81",
			commandLine: "",
			actions:     "Install",
			altered:     "4",
		},
		{
			name:        "action with two words",
			line:        "    79 | update                   | 2026-08-26 18:03 | I, U           |   24   ",
			id:          "79",
			commandLine: "update",
			actions:     "I, U",
			altered:     "24",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := reHistoryLine.FindStringSubmatch(tt.line)
			if matches == nil {
				t.Fatalf("line did not match: %q", tt.line)
			}
			for _, f := range []struct {
				field string
				got   string
				want  string
			}{
				{"transaction id", matches[1], tt.id},
				{"command line", matches[2], tt.commandLine},
				{"actions", matches[4], tt.actions},
				{"altered", matches[5], tt.altered},
			} {
				if f.got != f.want {
					t.Errorf("%s = %q, want %q", f.field, f.got, f.want)
				}
			}
		})
	}
}

// TestParseHistoryLineIgnoresHeader makes sure the separator row of
// 'dnf history list' is not mistaken for a transaction.
func TestParseHistoryLineIgnoresHeader(t *testing.T) {
	for _, line := range []string{
		"ID     | Command line             | Date and time    | Action(s)      | Altered",
		"-------------------------------------------------------------------------------",
		"",
	} {
		if matches := reHistoryLine.FindStringSubmatch(line); matches != nil {
			t.Errorf("line %q should not match, got %q", line, matches[0])
		}
	}
}

// TestParseAltered covers the status flags dnf appends to the Altered column,
// taken from the real history of a long lived host.
func TestParseAltered(t *testing.T) {
	tests := map[string]string{
		"   1   ":   "1",
		"  22   ":   "22",
		" 1295 EE ": "1295",
		"   1 >":    "1",
		"   2 ><":   "2",
		"   4 *":    "4",
		"   7 #":    "7",
		"":          "",
	}

	for input, want := range tests {
		if got := parseAltered(input); got != want {
			t.Errorf("parseAltered(%q) = %q, want %q", input, got, want)
		}
	}
}

// TestParseTransactionDetail uses an excerpt of a real interrupted transaction:
// dnf marks its items with "**" and they used to be dropped, leaving only the
// "Reason Change" lines, which were split into action "Reason" and a package
// named "Change kernel".
func TestParseTransactionDetail(t *testing.T) {
	out := `Transaction ID : 23
Begin time     : Tue 12 May 2026 03:14:26 PM -03
Begin rpmdb    : 7d6ead0ccf3d184b70d7b1a9ad4fd2cfb2ec6d1cd7e73d35926e615bbc508dee
End time       : Wed 31 Dec 1969 09:00:00 PM -03 (-1778609666 seconds)
End rpmdb      :
User           : System <unset>
Return-Code    : Failure: 1
Releasever     :
Command Line   : -y upgrade
Persistence    : Unknown
Comment        :
Packages Altered:
 ** Install       kernel-core-6.12.0-124.55.3.el10_1.x86_64                  @baseos
 ** Upgrade       libpng-2:1.6.40-8.el10_1.3.x86_64                          @baseos
 ** Upgraded      libpng-2:1.6.40-8.el10_1.2.x86_64                          @@System
    Reason Change kernel-core-6.12.0-124.31.1.el10_1.x86_64                  @baseos
 ** Removed       kernel-core-6.12.0-124.49.1.el10_1.x86_64                  @@System
    Install       vim-common-2:9.1.083-6.el10_1.4.x86_64                     @appstream
Scriptlet output:
   1 Job for zabbix-proxy.service canceled.
`
	d, err := parseTransactionDetail(out)
	if err != nil {
		t.Fatalf("parseTransactionDetail: %v", err)
	}

	want := []Package{
		{Action: "Install", Name: "kernel-core", Version: "6.12.0", Release: "124.55.3.el10_1", Arch: "x86_64", Repo: "baseos"},
		{Action: "Upgrade", Name: "libpng", Version: "1.6.40", Release: "8.el10_1.3", Epoch: "2", Arch: "x86_64", Repo: "baseos"},
		{Action: "Upgraded", Name: "libpng", Version: "1.6.40", Release: "8.el10_1.2", Epoch: "2", Arch: "x86_64", FromRepo: "System"},
		{Action: "Reason Change", Name: "kernel-core", Version: "6.12.0", Release: "124.31.1.el10_1", Arch: "x86_64", Repo: "baseos"},
		{Action: "Removed", Name: "kernel-core", Version: "6.12.0", Release: "124.49.1.el10_1", Arch: "x86_64", FromRepo: "System"},
		{Action: "Install", Name: "vim-common", Version: "9.1.083", Release: "6.el10_1.4", Epoch: "2", Arch: "x86_64", Repo: "appstream"},
	}
	if len(d.PackagesAltered) != len(want) {
		t.Fatalf("got %d packages, want %d: %+v", len(d.PackagesAltered), len(want), d.PackagesAltered)
	}
	for i, p := range d.PackagesAltered {
		if p != want[i] {
			t.Errorf("package %d: got %+v, want %+v", i, p, want[i])
		}
	}
	if d.ReturnCode != "Failure: 1" {
		t.Errorf("return code: got %q", d.ReturnCode)
	}
	if len(d.ScriptletOutput) != 1 || d.ScriptletOutput[0] != "1 Job for zabbix-proxy.service canceled." {
		t.Errorf("scriptlet output: got %q", d.ScriptletOutput)
	}
}
