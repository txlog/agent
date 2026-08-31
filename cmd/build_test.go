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
