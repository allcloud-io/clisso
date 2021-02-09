package aws

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-ini/ini"
)

func TestWriteToFile(t *testing.T) {
	id := "expiredkey"
	sec := "expiredsecret"
	tok := "expiredtoken"
	exp := time.Now()

	c := Credentials{
		AccessKeyID:     id,
		SecretAccessKey: sec,
		SessionToken:    tok,
		Expiration:      exp,
	}

	fn := "test_creds.txt"
	p := "expiredprofile"

	// Write credentials
	err := WriteToFile(&c, fn, p)
	if err != nil {
		t.Fatal("Could not write credentials to file: ", err)
	}

	// Sleep so above credentials expire
	time.Sleep(time.Duration(2) * time.Second)

	id = "testkey"
	sec = "testsecret"
	tok = "testtoken"
	exp = time.Now().Add(time.Duration(10) * time.Minute)

	c = Credentials{
		AccessKeyID:     id,
		SecretAccessKey: sec,
		SessionToken:    tok,
		Expiration:      exp,
	}

	p = "testprofile"

	// Write credentials
	err = WriteToFile(&c, fn, p)
	if err != nil {
		t.Fatal("Could not write credentials to file: ", err)
	}

	// Read credentials
	cfg, err := ini.Load(fn)
	if err != nil {
		t.Fatal("Could not load INI file: ", err)
	}

	p = "expiredprofile"
	s := cfg.Section(p)

	// Verify File
	keys := []string{"aws_access_key_id", "aws_secret_access_key", "aws_session_token", "aws_expiration"}
	for _, k := range keys {
		if s.HasKey(k) {
			t.Fatal("Expired profile was not cleaned up")
		}
	}

	p = "testprofile"
	s = cfg.Section(p)

	// Verify credentials
	k, err := s.GetKey("aws_access_key_id")
	if err != nil {
		t.Fatal(err)
	}
	if k.String() != id {
		t.Fatalf("Wrong access key ID: got %s, want %s", k.String(), id)
	}

	k, err = s.GetKey("aws_secret_access_key")
	if err != nil {
		t.Fatal(err)
	}
	if k.String() != sec {
		t.Fatalf("Wrong secret access key: got %s, want %s", k.String(), sec)
	}

	k, err = s.GetKey("aws_session_token")
	if err != nil {
		t.Fatal(err)
	}
	if k.String() != tok {
		t.Fatalf("Wrong session token: got %s, want %s", k.String(), tok)
	}

	k, err = s.GetKey("aws_expiration")
	if err != nil {
		t.Fatal(err)
	}
	f := exp.UTC().Format(time.RFC3339)
	if k.String() != f {
		t.Fatalf("Wrong expiration: got %s, want %s", k.String(), f)
	}

	err = os.Remove(fn)
	if err != nil {
		t.Fatalf("Could not remove file %v during cleanup", fn)
	}
}

func TestGetValidCredentials(t *testing.T) {
	fn := "test_creds.txt"

	id := "testkey"
	sec := "testsecret"
	tok := "testtoken"
	exp := time.Now()

	c := Credentials{
		AccessKeyID:     id,
		SecretAccessKey: sec,
		SessionToken:    tok,
		Expiration:      exp,
	}

	// Above credentials are expiring this exact moment..
	p := "expired"

	// Write credentials
	err := WriteToFile(&c, fn, p)
	if err != nil {
		t.Fatal("Could not write credentials to file: ", err)
	}

	// Expire in 1 hour
	c.Expiration = time.Now().Add(time.Duration(1) * time.Hour)
	p = "valid"

	// Write credentials
	err = WriteToFile(&c, fn, p)
	if err != nil {
		t.Fatal("Could not write credentials to file: ", err)
	}

	// Read credentials
	cfg, err := ini.Load(fn)
	if err != nil {
		t.Fatal("Could not load INI file: ", err)
	}

	s := cfg.Section("not-there")

	// Verify File
	if s.HasKey("aws_access_key_id") {
		t.Fatal("Section 'not-there' exists")
	}

	s = cfg.Section("valid")
	if !s.HasKey("aws_access_key_id") {
		t.Fatal("Section 'valid' is missing")
	}

	s = cfg.Section("expired")
	if !s.HasKey("aws_access_key_id") {
		t.Fatal("Section 'valid' is missing")
	}

	time.Sleep(time.Duration(1) * time.Second)

	data, err := GetValidCredentials(fn)
	if err != nil {
		t.Fatal("Failed to get NonExpiredCredentials")
	}

	if len(data.Profiles) != 1 {
		t.Fatal("Got more than 1 expected credential set")
	}

	if data.Profiles[0].Name != "valid" {
		t.Fatal("Returned wrong profile name")
	}

	if data.Profiles[0].LifetimeLeft.Seconds() < 3597 || data.Profiles[0].LifetimeLeft.Seconds() > 3599 {
		// Lets factor in some slow time
		t.Fatal("Expiration is outside of expected scope")
	}

	err = os.Remove(fn)
	if err != nil {
		t.Fatalf("Could not remove file %v during cleanup", fn)
	}

	_, err = GetValidCredentials(fn)
	if err != nil {
		t.Fatal("Function did crash on missing file")
	}
}

func TestWriteToShellUnix(t *testing.T) {
	id := "testkey"
	sec := "testsecret"
	tok := "testtoken"
	exp := time.Now()

	c := Credentials{
		AccessKeyID:     id,
		SecretAccessKey: sec,
		SessionToken:    tok,
		Expiration:      exp,
	}
	var b bytes.Buffer

	WriteToShell(&c, false, &b)

	got := b.String()
	want := fmt.Sprintf(
		"export AWS_ACCESS_KEY_ID=%v\nexport AWS_SECRET_ACCESS_KEY=%v\nexport AWS_SESSION_TOKEN=%v\n",
		id,
		sec,
		tok,
	)

	if got != want {
		t.Fatalf("Wrong info written to shell: got %v want %v", got, want)
	}
}

func TestWriteToShellWindows(t *testing.T) {
	id := "testkey"
	sec := "testsecret"
	tok := "testtoken"
	exp := time.Now()

	c := Credentials{
		AccessKeyID:     id,
		SecretAccessKey: sec,
		SessionToken:    tok,
		Expiration:      exp,
	}
	var b bytes.Buffer

	WriteToShell(&c, true, &b)

	got := b.String()
	want := fmt.Sprintf(
		"set AWS_ACCESS_KEY_ID=%v\nset AWS_SECRET_ACCESS_KEY=%v\nset AWS_SESSION_TOKEN=%v\n",
		id,
		sec,
		tok,
	)

	if got != want {
		t.Fatalf("Wrong info written to shell: got %v want %v", got, want)
	}
}
