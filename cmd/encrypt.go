package cmd

import (
	"github.com/spf13/cobra"
)

// encryptCmd represents the encrypt command
var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt a file or a variable for being encrypted",
	Long: `Encrypt a file or a variable for being encrypted
	encryption/decryption key can be provided with the --key flag
	or into [env-key-prefix]_VAULT_PASSWORD_EXEC env var, if it's a file and executable , ot will be executed to get the key
	or into [env-key-prefix]_VAULT_PASSWORD_FILE env var, if it'snt a file then it's taken as the key
	`,
	Run: func(cmd *cobra.Command, args []string) {
		Edit(cmd, args, false)
	},
}

func init() {
	rootCmd.AddCommand(encryptCmd)
}
