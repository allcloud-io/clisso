package keychain

import (
	"fmt"

	"github.com/challarao/keyring"
	"github.com/howeyc/gopass"
)

const (
	// KeyChainName is the name of the keychain used to store
	// passwords
	KeyChainName = "clisso"
)

// Keychain provides an interface to allow for the easy testing
// of this package
type Keychain interface {
	Get(string) ([]byte, error)
	Set(string, []byte) error
}

// DefaultKeychain provides a wrapper around github.com/tmc/keyring
// and provides defaults and abstractions for clisso to get passwords
type DefaultKeychain struct{}

// Set takes a provider in an argument, and a password from STDIN, and
// sets it in a keychain, should one exist.
func (DefaultKeychain) Set(provider string, password []byte) (err error) {
	return set(provider, password)
}

// Get will, once given a valid provider, return the password associated
// in order for logins to happen.
// If any error occours while talking to the keychain provider, we silently swallow it
// and just ask the user for the password instead. Error could be anything from access denied to
// password not found.
func (DefaultKeychain) Get(provider string) (pw []byte, err error) {
	pass, err := get(provider)
	if err != nil {
		// If we ever implement a logfile we might want to log what error occurred.
		fmt.Printf("Please enter %s password: ", provider)
		pass, err = gopass.GetPasswd()
		if err != nil {
			return nil, fmt.Errorf("couldn't read password from terminal")
		}
	}
	return pass, nil
}

func set(provider string, password []byte) (err error) {
	return keyring.Set(KeyChainName, provider, string(password))
}

func get(provider string) (pw []byte, err error) {
	pwString, err := keyring.Get(KeyChainName, provider)
	pw = []byte(pwString)
	return
}
