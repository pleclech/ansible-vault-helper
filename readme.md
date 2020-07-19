# ansible-vault-helper

A cli helper to read/write Ansible Vault secrets

## Description

avh can be used to edit/encrypt/decrypt Ansible Vault secrets given a key taken from cli flag or env variable.

input can be a file or stdin

output can be a file or stdout

## Help
avh --help

```
Helper to edit/encrypt/decrypt ansible vault file

Usage:
  avh [command]

Available Commands:
  decrypt     Decrypt file or var
  edit        Edit a file or a variable for being encrypted
  encrypt     Encrypt a file or a variable for being encrypted
  help        Help about any command
  version     show avh version

Flags:
  -d, --do-not-ask-for-key      even if key is not found for cli or env variable do not ask for it
  -e, --env-key-prefix string   prefix to be add in front of env var _VAULT_PASSWORD_EXEC or _VAULT_PASSWORD_FILE (default "DEFAULT")
  -h, --help                    help for avh
  -i, --input string            input file to edit or - to edit from stdin
  -k, --key string              raw encryption/decryption key
      --key-exec string         encryption/decryption key taken from an executable file
      --key-file string         encryption/decryption key taken from a file to encrypt/decrypt data
  -p, --key-prompt string       key prompt to show when asking for key (default "")
  -o, --output string           output file to save or - to print to stdout

Use "avh [command] --help" for more information about a command.
```

## Options

### setting encryption/decryption key:

key will be check in order:
cli flags take precedence over env variable

1. --key-exec or env variable [--env-prefix]_VAULT_PASSWORD_EXEC, the value is an executable file that can provide the key
2. --key-file or env variable [--env-prefix]_VAULT_PASSWORD_FILE, the file content is the key
3. --key or env variable [--env-prefix]_VAULT_PASSWORD, the value is the key

If the key is not found and --do-not-ask-for-key is not set then the key will be ask to be entered

### input

-i [format] if format is - then input is read from stdin otherwise from a file

### output

-o [format] if format is - then output to stdout otherwise to a file

## Edition

avh edit [options]

It will use the editor set in env variable EDITOR, if not set it will use per default :

- nano on linux
- notepad on windows

to use vscode for example set EDITOR to 'code'

If editing a content encrypted it will be decrypted before using the given key

After edition the content will be encrypted using the given key

## Examples

### edit a file to be encrypted 
avh edit -i my-file

### edit a var and show the encrypted value on stdout
echo "my secret content" | avh -i -

### edit a file given a key from cli and show output on stdout
avh edit -k "my secret key" -i my-file -o -

Same can be done for the other command encrypt / decrypt

## Doc

Check out the Ansible documentation regarding the Vault file format:

- https://docs.ansible.com/ansible/2.4/vault.html#vault-format

### Credits

Inspired from : 

- https://github.com/sosedoff/ansible-vault-go
- https://samrapdev.com/capturing-sensitive-input-with-editor-in-golang-from-the-cli/

## License

MIT
