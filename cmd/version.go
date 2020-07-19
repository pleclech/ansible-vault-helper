package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show avh version",
	Long:  `show avh version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("v%s", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
