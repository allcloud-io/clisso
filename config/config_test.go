package config

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

var testConfig = []byte(`
providers:
  sample-onelogin-provider:
    client-id: fake-id
    client-secret: fake-secret
    duration: 3600
    region: US
    subdomain: fake-domain
    type: onelogin
    username: fake@fake.com
  sample-okta-provider:
    base-url: https://fake.okta.com
    duration: 14400
    type: okta
    username: fake@fake.com
`)

func TestGetOneLoginProvider(t *testing.T) {
	c := newTestConfig(testConfig)

	want := OneLoginProvider{
		ClientID:     "fake-id",
		ClientSecret: "fake-secret",
		Duration:     3600,
		Region:       "US",
		Subdomain:    "fake-domain",
		Type:         "onelogin",
		Username:     "fake@fake.com",
	}

	p, err := c.GetOneLoginProvider("sample-onelogin-provider")
	if err != nil {
		t.Fatalf("Error getting provider: %v", err)
	}

	if !reflect.DeepEqual(*p, want) {
		t.Fatalf("Wrong provider returned: got %v, want %v", p, want)
	}
}

func TestGetOktaProvider(t *testing.T) {
	c := newTestConfig(testConfig)

	want := OktaProvider{
		BaseURL:  "https://fake.okta.com",
		Duration: 14400,
		Type:     "okta",
		Username: "fake@fake.com",
	}

	p, err := c.GetOktaProvider("sample-okta-provider")
	if err != nil {
		t.Fatalf("Error getting provider: %v", err)
	}

	if !reflect.DeepEqual(*p, want) {
		t.Fatalf("Wrong provider returned: got %v, want %v", p, want)
	}
}

func newTestConfig(b []byte) *Config {
	c := New()
	err := c.v.ReadConfig(bytes.NewBuffer(testConfig))
	if err != nil {
		panic(fmt.Sprintf("Error reading test config: %v", err))
	}

	return c
}
