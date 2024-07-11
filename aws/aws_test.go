/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package aws

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/allcloud-io/clisso/log"
	"github.com/go-ini/ini"
	"github.com/stretchr/testify/assert"
)

var _, hook = log.SetupLogger("panic", "", false, true)

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

	fn := "TestWriteToFile.txt"
	inifile := ini.Empty()
	// add some other config options to ensure we don't overwrite them
	// key value pairs
	test := map[string]string{
		"region": "us-west-2",
		"output": "json",
	}
	for k, v := range test {
		inifile.Section("expiredprofile_with_config").Key(k).SetValue(v)
	}
	err := inifile.SaveTo(fn)
	if err != nil {
		t.Fatal("Could not write INI file: ", err)
	}

	for _, p := range []string{"expiredprofile", "expiredprofile_with_config"} {
		// Write credentials
		err := OutputFile(&c, fn, p)
		if err != nil {
			t.Fatal("Could not write credentials to file: ", err)
		}
	}

	// Sleep so above credentials expire, we can't fake it since WriteToFile should not write expired credentials but have to wait.
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

	p := "testprofile"

	// Write credentials
	err = OutputFile(&c, fn, p)
	if err != nil {
		t.Fatal("Could not write credentials to file: ", err)
	}

	// Read credentials
	cfg, err := ini.Load(fn)
	if err != nil {
		t.Fatal("Could not load INI file: ", err)
	}

	p = "expiredprofile"
	if cfg.HasSection(p) {
		t.Fatalf("Expired profile was not cleaned up")
	}

	p = "expiredprofile_with_config"
	if !cfg.HasSection(p) {
		t.Fatalf("Expired profile with config was cleaned up")
	}

	s := cfg.Section(p)

	// Verify File
	keys := []string{"aws_access_key_id", "aws_secret_access_key", "aws_session_token", "aws_expiration"}
	for _, k := range keys {
		if s.HasKey(k) {
			t.Fatal("expiredprofile_with_config was not cleaned up correctly")
		}
	}

	// Verify other sections
	for k, v := range test {
		s = cfg.Section(p)
		k, err := s.GetKey(k)
		if err != nil {
			t.Fatal(err)
		}
		if k.String() != v {
			t.Fatalf("Wrong %s: got %s, want %s", k, k.String(), v)
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

func initConfig(filename string) error {
	inifile := ini.Empty()
	// add some other config options to ensure we don't overwrite them
	inifile.Section("default").Key("region").SetValue("us-west-2")
	inifile.Section("default").Key("output").SetValue("json")

	inifile.Section("cred-process").Key("credential_process").SetValue("echo")

	// profile setup for using a source profile
	inifile.Section("child-profile").Key("source-profile").SetValue("cred-process")
	inifile.Section("child-profile").Key("role_arn").SetValue("arn:aws:iam::123456789012:role/role-name")

	// mock an expired clisso temporary profile
	inifile.Section("expiredprofile").Key("aws_access_key_id").SetValue("expiredkey")
	inifile.Section("expiredprofile").Key("aws_secret_access_key").SetValue("expired")
	inifile.Section("expiredprofile").Key("aws_session_token").SetValue("expiredtoken")
	inifile.Section("expiredprofile").Key("aws_expiration").SetValue(time.Now().Add(-time.Duration(1) * time.Hour).UTC().Format(time.RFC3339))

	// mock a valid clisso temporary profile
	inifile.Section("validprofile").Key("aws_access_key_id").SetValue("testkey")
	inifile.Section("validprofile").Key("aws_secret_access_key").SetValue("testsecret")
	inifile.Section("validprofile").Key("aws_session_token").SetValue("testtoken")
	inifile.Section("validprofile").Key("aws_expiration").SetValue(time.Now().Add(time.Duration(1) * time.Hour).UTC().Format(time.RFC3339))

	return inifile.SaveTo(filename)

}
func TestProtectSections(t *testing.T) {
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

	fn := "TestProtectSections.txt"
	err := initConfig(fn)

	if err != nil {
		t.Fatal("Could not write INI file: ", err)
	}
	err = OutputFile(&c, fn, "default")
	if err != nil {
		t.Fatal("Could not write credentials to file: ", err)
	}

	for _, p := range []string{"cred-process", "child-profile"} {
		err = OutputFile(&c, fn, p)

		if err == nil {
			t.Fatalf("Write to %s should have been aborted", p)
		}
	}

	cfg, err := ini.Load(fn)
	if err != nil {
		t.Fatal("Could not load INI file: ", err)
	}
	// Verify other sections
	s := cfg.Section("default")
	k, err := s.GetKey("region")
	if err != nil {
		t.Fatal(err)
	}
	if k.String() != "us-west-2" {
		t.Fatalf("Wrong region: got %s, want us-west-2", k.String())
	}

	k, err = s.GetKey("output")
	if err != nil {
		t.Fatal(err)
	}
	if k.String() != "json" {
		t.Fatalf("Wrong output: got %s, want json", k.String())
	}

	s = cfg.Section("cred-process")
	k, err = s.GetKey("credential_process")
	if err != nil {
		t.Fatal(err)
	}
	if k.String() != "echo" {
		t.Fatalf("Wrong credential_process: got %s, want echo", k.String())
	}

	s = cfg.Section("child-profile")
	k, err = s.GetKey("source-profile")
	if err != nil {
		t.Fatal(err)
	}
	if k.String() != "cred-process" {
		t.Fatalf("Wrong source-profile: got %s, want cred-process", k.String())
	}

	k, err = s.GetKey("role_arn")
	if err != nil {
		t.Fatal(err)
	}
	if k.String() != "arn:aws:iam::123456789012:role/role-name" {
		t.Fatalf("Wrong role_arn: got %s, want arn:aws:iam::123456789012:role/role-name", k.String())
	}

	err = os.Remove(fn)
	if err != nil {
		t.Fatalf("Could not remove file %v during cleanup", fn)
	}
}

func TestGetValidProfiles(t *testing.T) {
	fn := "TestGetValidProfiles.txt"

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
	err := OutputFile(&c, fn, p)
	if err != nil {
		t.Fatal("Could not write credentials to file: ", err)
	}

	// Expire in 1 hour
	c.Expiration = time.Now().Add(time.Duration(1) * time.Hour)
	p = "valid"

	// Write credentials
	err = OutputFile(&c, fn, p)
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

	profiles, err := GetValidProfiles(fn)
	if err != nil {
		t.Fatal("Failed to get NonExpiredCredentials")
	}

	if len(profiles) != 1 {
		t.Fatal("Got more than 1 expected credential set")
	}

	if profiles[0].Name != "valid" {
		t.Fatal("Returned wrong profile name")
	}

	if profiles[0].LifetimeLeft.Seconds() < 3597 || profiles[0].LifetimeLeft.Seconds() > 3599 {
		// Lets factor in some slow time
		t.Fatal("Expiration is outside of expected scope")
	}

	err = os.Remove(fn)
	if err != nil {
		t.Fatalf("Could not remove file %v during cleanup", fn)
	}

	_, err = GetValidProfiles(fn)
	if err != nil {
		t.Fatal("Function did crash on missing file")
	}
}

func TestOutputUnixEnvironment(t *testing.T) {
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

	OutputEnvironment(&c, false, &b)

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

func TestOutputWindowsEnvironment(t *testing.T) {
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

	OutputEnvironment(&c, true, &b)

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

func TestSetCredentialProcess(t *testing.T) {
	assert := assert.New(t)
	fn := "TestSetCredentialProcess.txt"
	err := initConfig(fn)
	assert.Nil(err, "Expected no error, but got: %v", err)

	// nothing to todo, should be skipped
	p := "cred-process"
	err = SetCredentialProcess(fn, p)
	assert.Nil(err, "Expected no error, but got: %v", err)
	assert.GreaterOrEqual(len(hook.Entries), 1, "Expected 1 or more log message, but got: %v", hook.Entries)
	expected := fmt.Sprintf(infoProfileAlreadyConfigured, p)
	assert.Equal(expected, hook.LastEntry().Message, "Expected '%s', but got: %v", expected, hook.LastEntry().Message)
	hook.Reset()

	// set credential process on child-profile should fail
	err = SetCredentialProcess(fn, "child-profile")
	assert.NotNil(err, "Expected an error, but got nil")
	assert.GreaterOrEqual(len(hook.Entries), 1, "Expected 1 or more log message, but got: %v", hook.Entries)
	expected = fmt.Sprintf(errCannotBeUsed, "child-profile", "role_arn")
	assert.Equal(expected, err.Error(), "Expected '%s', but got: %v",expected, err.Error())
	hook.Reset()

	// set credential process on expired and valid profile should work and remove the profile keys
	for _, p := range []string{"expiredprofile", "validprofile"} {
		err = SetCredentialProcess(fn, p)
		assert.Nil(err, "Expected no error, but got: %v", err)
		assert.GreaterOrEqual(len(hook.Entries), 1, "Expected 1 or more log message, but got: %v", hook.Entries)
		expected = fmt.Sprintf(infoProfileConfigured, p)
		assert.Equalf(expected, hook.LastEntry().Message, "Expected '%s', but got: %v", expected, hook.LastEntry().Message)
		hook.Reset()

		cfg, err := ini.Load(fn)
		assert.Nil(err, "Expected no error, but got: %v", err)
		// the section should only have on key left
		s := cfg.Section(p)
		assert.Equal(1, len(s.Keys()), "Expected 1 key, but got: %v", len(s.Keys()))
		// the key should be credential_process
		k, err := s.GetKey("credential_process")
		assert.Nil(err, "Expected no error, but got: %v", err)
		expected = fmt.Sprintf(credentialProcessFormat, p)
		assert.Equal(expected, k.String(), "Expected '%s', but got: %v", expected, k.String())
	}

	os.Remove(fn)
}
