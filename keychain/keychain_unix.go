// +build !windows

package keychain

import (
	"github.com/tmc/keyring"
)

// DefaultKeychain provides a wrapper around github.com/tmc/keyring as well as defaults and
// abstractions for Clisso.
type DefaultKeychain struct{}

// Set stores the given password for the given provider in a keychain.
func (*DefaultKeychain) Set(provider string, password []byte) (err error) {
	return keyring.Set(KeyChainName, provider, string(password))
}

// Get returns the stored password for the given provider, or an error.
func (*DefaultKeychain) Get(provider string) (pw []byte, err error) {
	pwString, err := keyring.Get(KeyChainName, provider)
	pw = []byte(pwString)

	return
}

// NewKeychain returns a new DefaultKeychain.
func NewKeychain() *DefaultKeychain {
	return &DefaultKeychain{}
}
