package keychain

const (
	// KeyChainName is the name of the keychain used by Clisso to store passwords.
	KeyChainName = "clisso"
)

// Keychain provides an interface to allow for easy testing.
type Keychain interface {
	Get(string) ([]byte, error)
	Set(string, []byte) error
}
