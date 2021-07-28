package keychain

import (
	"math/rand"
	"testing"
	"time"
)

func randSeq(n int, letters []rune) []byte {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return []byte(string(b))
}

// TestCycle sets a password and checks if the password can be retrieved again
// we use a random password to make sure the set was really successful and not due to a previous run
func TestCycle(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	keyChain := DefaultKeychain{}

	for _, test := range []struct {
		name           string
		password           []byte
	}{
		{
			"clissotest-random",
			randSeq(20, []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")),
		},
		{
			"clissotest-umlaut",
			randSeq(10,[]rune("üöäßÜÖÄ")),
		},
		{
			"clissotest-greek",
			randSeq(10,[]rune("νβζγντ")),
		},
	}{
		t.Run(test.name, func(t *testing.T) {
			err := keyChain.Set(test.name, test.password)
			if err != nil {
				t.Errorf("unexpected error %+v", err)
			}

			retrievedPass, err := keyChain.Get(test.name)
			if err != nil {
				t.Errorf("unexpected error %+v", err)
			}

			if string(test.password) != string(retrievedPass) {
				t.Errorf("expected %s, received %s", test.password, retrievedPass)
			}
		})
	}
}
