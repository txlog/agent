package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/txlog/agent/util"
)

var copyfailCmd = &cobra.Command{
	Use:   "copyfail",
	Short: "Check if this host is vulnerable to CVE-2026-31431 (Copy Fail)",
	Long: `
Performs a safe, non-destructive test to determine if the running kernel
is vulnerable to CVE-2026-31431 (Copy Fail), a local privilege escalation
bug in the Linux kernel's authencesn cryptographic template.

The test exercises the same kernel code path used by the exploit but only
writes to a temporary file owned by the agent. No system files are modified.

Phase 1: Tests if the kernel allows page cache writes via AF_ALG + authencesn
Phase 2: Verifies if privilege escalation to root is possible (checks setuid
         binaries exist and are reachable via the exploit pipeline)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, strings.Repeat("=", 60))
		fmt.Fprintln(os.Stdout, "🔍 CVE-2026-31431 (Copy Fail) — Vulnerability Check")
		fmt.Fprintln(os.Stdout, strings.Repeat("=", 60))
		fmt.Fprintln(os.Stdout)

		result := util.CheckCopyFail()

		// Print detailed step-by-step results
		for _, line := range strings.Split(result.Details, "\n") {
			if line == "" {
				continue
			}
			// Red: vulnerability confirmed, escalation conditions met
			if strings.Contains(line, "VULNERABLE") ||
				strings.Contains(line, "DETECTED") ||
				strings.Contains(line, "escalation conditions MET") {
				fmt.Fprintf(os.Stdout, "   ✗ %s\n", color.RedString(line))
				// Yellow: errors, failures, inconclusive
			} else if strings.Contains(line, "error") ||
				strings.Contains(line, "failed") ||
				strings.Contains(line, "unlikely") ||
				strings.Contains(line, "Inconclusive") {
				fmt.Fprintf(os.Stdout, "   ⚠ %s\n", color.YellowString(line))
				// Green: everything else (positive findings, neutral info)
			} else {
				fmt.Fprintf(os.Stdout, "   ✓ %s\n", color.GreenString(line))
			}
		}

		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, strings.Repeat("=", 60))

		if !result.Vulnerable {
			color.Green("RESULT: NOT VULNERABLE")
			fmt.Fprintln(os.Stdout, strings.Repeat("=", 60))
			fmt.Fprintf(os.Stdout, "   CVE:     %s\n", color.CyanString("CVE-2026-31431"))
			fmt.Fprintf(os.Stdout, "   Status:  %s\n", color.GreenString(result.Description))
		} else if result.EscalationConfirmed {
			color.Red("RESULT: VULNERABLE — Privilege escalation confirmed")
			fmt.Fprintln(os.Stdout, strings.Repeat("=", 60))
			fmt.Fprintf(os.Stdout, "   CVE:     %s\n", color.CyanString("CVE-2026-31431"))
			fmt.Fprintf(os.Stdout, "   Target:  %s\n", color.RedString(result.SetuidTarget))
			fmt.Fprintf(os.Stdout, "   Impact:  %s\n", color.RedString("Any unprivileged user can gain root access"))
			fmt.Fprintf(os.Stdout, "   Fix:     %s\n", color.CyanString("Update kernel to include commit a664bf3d603d"))
			fmt.Fprintln(os.Stdout)
			fmt.Fprintf(os.Stdout, "   Workaround:\n")
			fmt.Fprintf(os.Stdout, "     %s\n", color.YellowString("echo \"install algif_aead /bin/false\" > \\"))
			fmt.Fprintf(os.Stdout, "     %s\n", color.YellowString("  /etc/modprobe.d/disable-algif-aead.conf && \\"))
			fmt.Fprintf(os.Stdout, "     %s\n", color.YellowString("  rmmod algif_aead"))
		} else {
			color.Red("RESULT: VULNERABLE — Kernel bug confirmed")
			fmt.Fprintln(os.Stdout, strings.Repeat("=", 60))
			fmt.Fprintf(os.Stdout, "   CVE:     %s\n", color.CyanString("CVE-2026-31431"))
			fmt.Fprintf(os.Stdout, "   Status:  %s\n", color.YellowString(result.Description))
			fmt.Fprintf(os.Stdout, "   Fix:     %s\n", color.CyanString("Update kernel to include commit a664bf3d603d"))
		}

		fmt.Fprintln(os.Stdout, strings.Repeat("=", 60))
		fmt.Fprintln(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(copyfailCmd)
}
