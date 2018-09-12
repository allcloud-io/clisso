// +build windows

package keychain

import "errors"

// NoopKeychain represents a fake keychain which doesn't do anything. It is used instead of the
// a real keychain on systems without keychain support, i.e. Windows.
type NoopKeychain struct{}

// Get is a noop implementation of the Keychain interface's Get() function.
func (k *NoopKeychain) Get(string) ([]byte, error) {
	return []byte{}, errors.New("not implemented")
}

// Set is a noop implementation of the Keychain interface's Set() function.
func (k *NoopKeychain) Set(string, []byte) error {
	return errors.New("not implemented")
}

// NewNoopKeychain returns a new NoopKeychain.
func NewNoopKeychain() *NoopKeychain {
	return &NoopKeychain{}
}
