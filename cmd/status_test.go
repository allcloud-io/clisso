package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/allcloud-io/clisso/aws"
	"github.com/spf13/viper"
)

func TestPrintStatus(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "credentials")
	c := aws.Credentials{AccessKeyID: "k", SecretAccessKey: "s", SessionToken: "t", Expiration: time.Now().Add(time.Hour)}
	if err := aws.OutputFile(&c, fn, "valid-app"); err != nil {
		t.Fatal(err)
	}
	// write an expired profile directly (OutputFile cleans expired ones)
	f, _ := os.OpenFile(fn, os.O_APPEND|os.O_WRONLY, 0600)
	_, _ = f.WriteString("[expired-app]\naws_access_key_id = k\naws_secret_access_key = s\naws_session_token = t\naws_expiration = " +
		time.Now().Add(-time.Hour).UTC().Format(time.RFC3339) + "\n")
	f.Close()

	viper.Set("global.output", fn)
	defer viper.Set("global.output", "")

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	printStatus()
	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)

	s := string(out)
	if !strings.Contains(s, "valid-app") || strings.Contains(s, "expired-app") {
		t.Errorf("unexpected status output:\n%s", s)
	}
	if !strings.Contains(s, "APP") || !strings.Contains(s, "REMAINING") {
		t.Errorf("table header missing:\n%s", s)
	}
}
