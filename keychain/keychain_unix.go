// +build !windows

package keychain

import "github.com/tmc/keyring"

func set(provider string, password []byte) (err error) {
	return keyring.Set(KeyChainName, provider, string(password))
}

func get(provider string) (pw []byte, err error) {
	pwString, err := keyring.Get(KeyChainName, provider)
	pw = []byte(pwString)
	return
}
