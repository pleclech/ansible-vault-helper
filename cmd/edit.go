package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pleclech/ansible-vault-helper/editor"
	"github.com/pleclech/ansible-vault-helper/vault"

	"github.com/spf13/cobra"
)

var (
	reYamlVaultEntry = regexp.MustCompile(`([^\S\r\n]*)(\w+):(\s+)[!]vault(.*)[\||>].*`)
	reSpaces         = regexp.MustCompile(`(\s+)`)
)

type InputInfo struct {
	content    []byte
	isFile     bool
	fileMode   os.FileMode
	tmpFileExt string
	fileExt    string
	key        string
}

func (i *InputInfo) decryptYamlEntries() error {
	content := string(i.content)
	matches := reYamlVaultEntry.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		lMin := len(match[1])
		pat := fmt.Sprintf(`%s(((\s{%d,})(.+))+)`, regexp.QuoteMeta(match[0]), lMin+2)
		reN := regexp.MustCompile(pat)
		values := reN.FindAllStringSubmatch(content, -1)
		if len(values) > 0 {
			value := values[0][1]
			tmpValue := reSpaces.ReplaceAllString(value, "\n")[1:]
			if vault.MaybeEncrypted(tmpValue) {
				ic, err := vault.Decrypt(tmpValue, i.key)
				if err != nil {
					return err
				}
				sep := "\n" + strings.Repeat(" ", lMin+2)
				var bb bytes.Buffer
				for _, tmp := range strings.Split(ic, "\n") {
					bb.WriteString(sep)
					bb.WriteString(tmp)
				}
				content = strings.Replace(content, value, bb.String(), -1)
			}
		}
	}
	i.content = []byte(content)
	return nil
}

func (i InputInfo) IsYaml() bool {
	return i.tmpFileExt == ".yaml" || i.tmpFileExt == ".yml"
}

func (i *InputInfo) Encrypt() (string, error) {
	content := string(i.content)
	if !i.IsYaml() {
		return vault.Encrypt(content, i.key, 0)
	}
	matches := reYamlVaultEntry.FindAllStringSubmatch(content, -1)
	if len(matches) <= 0 {
		return vault.Encrypt(content, i.key, 0)
	}
	for _, match := range matches {
		lMin := len(match[1])
		pat := fmt.Sprintf(`%s(((\s{%d,})(.+))+)`, regexp.QuoteMeta(match[0]), lMin+2)
		reN := regexp.MustCompile(pat)
		values := reN.FindAllStringSubmatch(content, -1)
		if len(values) > 0 {
			value := values[0][1]
			var bb bytes.Buffer
			sep := ""
			for _, tmp := range strings.Split(value[1:], "\n") {
				tmp = strings.TrimSpace(tmp)
				bb.WriteString(sep)
				bb.WriteString(tmp)
				sep = "\n"
			}
			ic, err := vault.Encrypt(bb.String(), i.key, lMin+2)
			if err != nil {
				return content, err
			}
			content = strings.Replace(content, value, "\n"+ic, -1)
		}
	}
	return content, nil
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
		if i.isFile {
			if key == "" && !doNotAskForKey {
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
			switch i.tmpFileExt {
			case ".yaml", ".yml":
				return i.decryptYamlEntries()
			}
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
		var lines bytes.Buffer

		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					lines.WriteString(line)
				}
				break
			}
			lines.WriteString(line)
		}

		inputInfo.content = lines.Bytes()
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
		inputInfo.tmpFileExt = filepath.Ext(input[0 : len(input)-len(inputInfo.fileExt)])
		if inputInfo.tmpFileExt == "" {
			inputInfo.tmpFileExt = inputInfo.fileExt
		}
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

	ext := inputInfo.tmpFileExt

	err = inputInfo.Decrypt(doNotAskForKey, keyPrompt)
	if err != nil {
		panic(err)
	}

	var editedBytes []byte

	if openEditor {
		editedBytes, err = editor.CaptureInputFromEditor(
			editor.GetPreferredEditorFromEnvironment,
			inputInfo.content,
			ext,
		)
		if err != nil {
			panic(err)
		}
		inputInfo.content = editedBytes
	}

	encString, err := inputInfo.Encrypt()
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
