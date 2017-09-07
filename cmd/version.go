package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the logon command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version for this tool",
	Long:  `current version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(_VERSION)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
