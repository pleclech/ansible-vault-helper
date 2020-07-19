package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pleclech/ansible-vault-helper/cleanup"
	"github.com/pleclech/ansible-vault-helper/vault"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	version, envKeyPrefix, cfgFile, input, output, vaultKeyExec, vaultKeyFile, vaultKey, keyPrompt string
	doNotAskForKey                                                                                 bool
)

func GetKeyFromFlags() vault.Key {
	isExec := false
	isFile := false

	key := vaultKeyExec
	for {
		if key != "" {
			isFile = true
			isExec = true
			break
		}
		key = vaultKeyFile
		if key != "" {
			isFile = true
			break
		}
		key = vaultKey
		break
	}

	return vault.Key{
		Value:  key,
		IsExec: isExec,
		IsFile: isFile,
	}
}

func readPassword(label string, keyPrompt string) (string, error) {
	fd := int(os.Stdin.Fd())

	oldState, err := terminal.GetState(fd)
	if err != nil {
		return "", fmt.Errorf("could not get state of terminal : %w", err)
	}

	defer cleanup.Trap(
		func() {
			terminal.Restore(fd, oldState)
		},
	)()

	if keyPrompt != "" {
		keyPrompt = fmt.Sprintf("(%s)", keyPrompt)
	}

	fmt.Fprintf(os.Stderr, "%s %s: ", label, keyPrompt)
	tmp, err := terminal.ReadPassword(fd)
	fmt.Fprint(os.Stderr, "\n")
	if err != nil {
		return "", fmt.Errorf("%s : %w", label, err)
	}
	return string(tmp), nil
}

func writeToFile(fileName string, content string, mode os.FileMode) error {
	tmpName := fileName + ".tmp"

	err := ioutil.WriteFile(tmpName, ([]byte)(content), mode)
	if err != nil {
		return err
	}

	defer cleanup.Trap(func() { os.Remove(tmpName) })()

	return os.Rename(tmpName, fileName)
}

var rootCmd = &cobra.Command{
	Use:   "avh",
	Short: "Ansible Vault Helper",
	Long:  `Helper to edit/encrypt/decrypt ansible vault file`,
	// Run: func(cmd *cobra.Command, args []string) {
	// },
}

func Execute(v string) {
	version = v
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	pf := rootCmd.PersistentFlags()
	//	pf.StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.avh.yaml)")
	pf.StringVarP(&envKeyPrefix, "env-key-prefix", "e", "DEFAULT", "prefix to be add in front of env var _VAULT_PASSWORD_EXEC or _VAULT_PASSWORD_FILE")
	pf.StringVarP(&input, "input", "i", "", "input file to edit or - to edit from stdin")
	pf.StringVarP(&output, "output", "o", "", "output file to save or - to print to stdout")
	pf.StringVarP(&vaultKey, "key", "k", "", "raw encryption/decryption key")
	pf.StringVar(&vaultKeyExec, "key-exec", "", "encryption/decryption key taken from an executable file")
	pf.StringVar(&vaultKeyFile, "key-file", "", "encryption/decryption key taken from a file to encrypt/decrypt data")
	pf.BoolVarP(&doNotAskForKey, "do-not-ask-for-key", "d", false, "even if key is not found do not ask for it")
	pf.StringVarP(&keyPrompt, "key-prompt", "p", "", "key prompt to show when asking for key")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".avh")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
