package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/txlog/agent/util"
)

var failuresCmd = &cobra.Command{
	Use:   "failures",
	Short: "Check if this host is vulnerable to known kernel exploits",
	Long: `
Performs safe, non-destructive tests to determine if the running kernel
is vulnerable to known local privilege escalation exploits:

  - Copy Fail (CVE-2026-31431): AF_ALG authencesn page cache write
  - Dirty Frag: ESP/XFRM page cache write via seq_hi
  - Fragnesia: skb shared-frag coalescing page cache write`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, strings.Repeat("=", 60))
		fmt.Fprintln(os.Stdout, "🔍 Linux Kernel Vulnerability Assessment")
		fmt.Fprintln(os.Stdout, strings.Repeat("=", 60))
		fmt.Fprintln(os.Stdout)

		// Check Copy Fail
		copyFailResult := util.CheckCopyFail()
		printVulnLine("Copy Fail (CVE-2026-31431)", copyFailResult.Vulnerable)

		// Check Dirty Frag
		dirtyFragResult := util.CheckDirtyFrag()
		printVulnLine("Dirty Frag", dirtyFragResult.Vulnerable)

		// Check Fragnesia
		fragnesiaResult := util.CheckFragnesia()
		printVulnLine("Fragnesia", fragnesiaResult.Vulnerable)

		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, strings.Repeat("=", 60))
		fmt.Fprintln(os.Stdout)
	},
}

// printVulnLine prints a single vulnerability status line with color formatting.
func printVulnLine(name string, vulnerable bool) {
	padded := fmt.Sprintf("%-40s", name)
	if vulnerable {
		fmt.Fprintf(os.Stdout, "   ✗ %s %s\n", padded, color.RedString("VULNERABLE"))
	} else {
		fmt.Fprintf(os.Stdout, "   ✓ %s %s\n", padded, color.GreenString("NOT VULNERABLE"))
	}
}

func init() {
	rootCmd.AddCommand(failuresCmd)
}
