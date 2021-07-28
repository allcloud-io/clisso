package keychain

import "testing"

func TestCycle(t *testing.T) {
	provider := "clissotest"
	pass := []byte("MyPass")

	keyChain := DefaultKeychain{}
	keyChain.Set(provider, pass)
	retrievedPass, err := keyChain.Get(provider)
	if err != nil {
		t.Errorf("unexpected error %+v", err)
	}
	if (string(pass) != string(retrievedPass)) {
		t.Error("Password storage failed")
	}
}