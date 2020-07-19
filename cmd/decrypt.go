package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// decryptCmd represents the decrypt command
var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt file or var",
	Long: `Decrypt encrypt file or var
	encryption/decryption key can be provided with the --key flag
	or into [env-key-prefix]_VAULT_PASSWORD_EXEC env var, if it's a file and executable , ot will be executed to get the key
	or into [env-key-prefix]_VAULT_PASSWORD_FILE env var, if it'snt a file then it's taken as the key
	`,
	Run: func(cmd *cobra.Command, args []string) {
		inputInfo, err := GetInputInfo(input, GetKeyFromFlags(), envKeyPrefix)
		if err != nil {
			panic(err)
		}

		err = inputInfo.Decrypt(doNotAskForKey, keyPrompt)
		if err != nil {
			panic(err)
		}

		decString := string(inputInfo.content)

		switch output {
		case "", "-":
			fmt.Print(decString)
		default:
			if output == input {
				panic("saving decrypted file : output file can't be the same as input")
			}
			err = writeToFile(output, decString, inputInfo.fileMode)
			if err != nil {
				panic(fmt.Errorf("saving decrypted file : %w", err))
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(decryptCmd)
}
