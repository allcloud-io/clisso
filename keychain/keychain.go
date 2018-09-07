package keychain

import (
	"errors"
	"log"
	"runtime"

	"github.com/fatih/color"
	"github.com/tmc/keyring"
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
	if runtime.GOOS == "windows" {
		log.Fatal(color.RedString("Storing passwords is not supported on windows"))
	}
	return keyring.Set(KeyChainName, provider, string(password))
}

// Get will, once given a valid provider, return the password associated
// in order for logins to happen
func (DefaultKeychain) Get(provider string) (pw []byte, err error) {
	if runtime.GOOS == "windows" {
		return nil, errors.New("Platform is not supported yet")
	}
	pwString, err := keyring.Get(KeyChainName, provider)
	pw = []byte(pwString)

	return
}
