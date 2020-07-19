package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pleclech/ansible-vault-helper/editor"
	"github.com/pleclech/ansible-vault-helper/vault"

	"github.com/spf13/cobra"
)

type InputInfo struct {
	content  []byte
	isFile   bool
	fileMode os.FileMode
	fileExt  string
	key      string
}

func (i *InputInfo) Decrypt(doNotAskForKey bool, keyPrompt string) error {
	var err error

	key := i.key
	ic := string(i.content)
	if vault.MaybeEncrypted(ic) {
		if key == "" && !doNotAskForKey {
			key, err = readPassword("Enter key", keyPrompt)
			if err != nil {
				return err
			}
			i.key = key
		}
		ic, err = vault.Decrypt(ic, key)
		if err != nil {
			return err
		}
		i.content = ([]byte)(ic)
	} else {
		if i.isFile && key == "" && !doNotAskForKey {
			key, err = readPassword("Enter new key", keyPrompt)
			if err != nil {
				return err
			}

			key2, err := readPassword("Confirm new key", keyPrompt)
			if err != nil {
				return err
			}

			if key != key2 {
				return fmt.Errorf("error password differs")
			}
			i.key = key
		}
	}
	return nil
}

func GetInputInfo(input string, keyChoice vault.Key, envKeyPrefix string) (*InputInfo, error) {
	inputInfo := &InputInfo{}

	key, err := vault.GetKey(keyChoice, envKeyPrefix)
	if err != nil {
		panic(err)
	}

	inputInfo.key = key

	switch input {
	case "", "-":
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		inputInfo.content = scanner.Bytes()
	default:
		stat, err := os.Stat(input)

		inputInfo.isFile = true

		if os.IsNotExist(err) {
			inputInfo.fileMode = vault.DefaultFileMode
		} else {
			inputInfo.fileMode = stat.Mode()
			inputInfo.content, err = ioutil.ReadFile(input)
			if err != nil {
				return inputInfo, err
			}
		}

		inputInfo.fileExt = filepath.Ext(input)
	}

	return inputInfo, nil
}

func Edit(cmd *cobra.Command, args []string, openEditor bool) {
	inputInfo, err := GetInputInfo(input, GetKeyFromFlags(), envKeyPrefix)
	if err != nil {
		panic(err)
	}

	if inputInfo.isFile && output == "" {
		output = input
	}

	err = inputInfo.Decrypt(doNotAskForKey, keyPrompt)
	if err != nil {
		panic(err)
	}

	var editedBytes []byte

	if openEditor {

		editedBytes, err = editor.CaptureInputFromEditor(
			editor.GetPreferredEditorFromEnvironment,
			inputInfo.content,
			inputInfo.fileExt,
		)
		if err != nil {
			panic(err)
		}
	}

	encString, err := vault.Encrypt(string(editedBytes), inputInfo.key)
	if err != nil {
		panic(fmt.Errorf("encrypt : %w", err))
	}

	switch output {
	case "", "-":
		fmt.Print(encString)
	default:
		err = writeToFile(output, encString, inputInfo.fileMode)
		if err != nil {
			panic(fmt.Errorf("saving encrypted file : %w", err))
		}
	}
}

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a file or a variable for being encrypted",
	Long: `Edit a file or a variable for being encrypted opening default editor in env variable EDITOR
	encryption/decryption key can be provided with the --key flag
	or into [env-key-prefix]_VAULT_PASSWORD_EXEC env var, if it's a file and executable , ot will be executed to get the key
	or into [env-key-prefix]_VAULT_PASSWORD_FILE env var, if it'snt a file then it's taken as the key
	`,
	Run: func(cmd *cobra.Command, args []string) {
		Edit(cmd, args, true)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
