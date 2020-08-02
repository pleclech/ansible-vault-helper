package vault

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const (
	DefaultFileMode = 0666
	execModeAll     = 0111
	envKeyExec      = "_VAULT_PASSWORD_EXEC"
	envKeyFile      = "_VAULT_PASSWORD_FILE"
	envKey          = "_VAULT_PASSWORD"
)

var (
	// ErrEmptyPassword is returned when password is empty
	ErrEmptyPassword = errors.New("password is blank")

	// ErrInvalidFormat is returned when secret content is not valid
	ErrInvalidFormat = errors.New("invalid secret format")

	// ErrInvalidPadding is returned when invalid key is used
	ErrInvalidPadding = errors.New("invalid padding")

	ErrKeyFileNotExec = errors.New("key file is not executable")

	ErrKeyFileNotFound = errors.New("key file not found")
)

// Encrypt encrypts the input string with the vault password
func Encrypt(input string, password string, pad int) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	salt, err := generateRandomBytes(saltLength)
	if err != nil {
		return "", err
	}
	key := generateKey([]byte(password), salt)

	// Encrypt the secret content
	data, err := encrypt([]byte(input), salt, key)
	if err != nil {
		return "", err
	}

	// Hash the secret content
	hash := hmac.New(sha256.New, key.hmacKey)
	hash.Write(data)
	hashSum := hash.Sum(nil)

	// Encode the secret payload
	return encodeSecret(&secret{data: data, salt: salt, hmac: hashSum}, key, pad)
}

// EncryptFile encrypts the input string and saves it into the file
func EncryptFile(path string, input string, password string) error {
	result, err := Encrypt(input, password, 0)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, []byte(result), DefaultFileMode)
}

// Return true if the input maybe encrypted as an ansible vault
func MaybeEncrypted(input string) bool {
	scanner := bufio.NewScanner(strings.NewReader(input))

	if !scanner.Scan() {
		return false
	}

	if strings.TrimSpace(scanner.Text()) != vaultHeader {
		return false
	}

	return true
}

// Decrypt decrypts the input string with the vault password
func Decrypt(input string, password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	lines := strings.Split(input, "\n")

	// Valid secret must include header and body
	if len(lines) < 2 {
		return "", ErrInvalidFormat
	}

	// Validate the vault file format
	if strings.TrimSpace(lines[0]) != vaultHeader {
		return "", ErrInvalidFormat
	}

	decoded, err := hex.DecodeString(strings.Join(lines[1:], ""))
	if err != nil {
		return "", err
	}

	secret, err := decodeSecret(string(decoded))
	if err != nil {
		return "", err
	}

	key := generateKey([]byte(password), secret.salt)
	if err := checkDigest(secret, key); err != nil {
		return "", err
	}

	result, err := decrypt(secret, key)
	if err != nil {
		return "", err
	}

	return result, nil
}

// DecryptFile decrypts the content of the file with the vault password
func DecryptFile(path string, password string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return Decrypt(string(data), password)
}

// obtain a key from a file, if run is true the file will be executed
func GetKeyFromFile(fileName string, run bool) (string, error) {
	stat, err := os.Stat(fileName)
	if err != nil {
		return fileName, err
	}

	if run {
		if (stat.Mode() & execModeAll) == 0 {
			return fileName, ErrKeyFileNotExec
		}
		var stdout bytes.Buffer
		cmd := exec.Command(fileName)
		cmd.Stdin = os.Stdin
		cmd.Stdout = &stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return fileName, err
		}
		return string(stdout.Bytes()), nil
	}

	key, err := ioutil.ReadFile(fileName)
	return string(key), err
}

// get key from specified key or from env var if not specified
func GetKey(keyChoice Key, envPrefix string) (string, error) {
	key := keyChoice.Value
	isExec := keyChoice.IsExec
	isFile := keyChoice.IsFile

	label := ""

	if key != "" {
		if !isFile {
			return key, nil
		}
		switch {
		case isExec:
			label = "from --key-exec"
		case isFile:
			label = "from --key-file"
		}
	} else {
		for {
			label = "from env"
			envName := envPrefix + envKeyExec
			key = os.Getenv(envName)
			if key != "" {
				label = fmt.Sprintf("%s %s", label, envName)
				isExec = true
				isFile = true
				break
			}

			envName = envPrefix + envKeyFile
			key = os.Getenv(envName)
			if key != "" {
				label = fmt.Sprintf("%s %s", label, envName)
				isFile = true
				break
			}

			envName = envPrefix + envKey
			key = os.Getenv(envName)
			break
		}
	}

	var err error

	if isFile {
		key, err = GetKeyFromFile(key, isExec)
		if err != nil {
			return key, fmt.Errorf("%s : %w", label, err)
		}
	}

	return key, err
}
