/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package keychain

import (
	"fmt"
	"syscall"

	log "github.com/sirupsen/logrus"
	keyring "github.com/zalando/go-keyring"
	"golang.org/x/term"
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
	log.WithField("provider", provider).Trace("Reading password from keychain")
	pass, err := get(provider)
	if err != nil {
		log.WithError(err).Trace("Couldn't read password from keychain")
		fmt.Printf("Please enter %s password: ", provider)
		pass, err = term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			err = fmt.Errorf("couldn't read password from terminal: %w", err)
			log.WithError(err).Trace("Couldn't read password from terminal")
			return nil, err
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
