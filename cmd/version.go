package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const txlogVersion = "0.1.1"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(txlogVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
