package saml

import (
	"io/ioutil"
	"testing"
)

func TestDecode(t *testing.T) {
	for _, test := range []struct {
		name        string
		path        string
		expectError bool
	}{
		{"Happy path, valid saml", "testdata/valid-response", false},
		{"Bad XML", "testdata/invalid-response", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			b, _ := ioutil.ReadFile(test.path)

			_, err := decode(string(b))
			if test.expectError && err == nil {
				t.Errorf("expected error")
			}
			if !test.expectError && err != nil {
				t.Errorf("unexpected error %+v", err)
			}
		})
	}
}

func TestGet(t *testing.T) {
	for _, test := range []struct {
		name        string
		path        string
		expectRole  string
		expectError bool
	}{
		{"Single ARN", "testdata/single-arn-response", "arn:aws:iam::123456789012:role/OneLogin-MyRole", false},
		//{"Many ARNs", "testdata/valid-response", "", false},         // will ask questions
		{"No ARNs", "testdata/no-arns-resonse", "", true},
		{"No ARN value", "testdata/no-arn-value-response", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			b, _ := ioutil.ReadFile(test.path)

			arn, err := Get(string(b))
			if test.expectError && err == nil {
				t.Errorf("expected error")
			}
			if !test.expectError && err != nil {
				t.Errorf("unexpected error %+v", err)
			}

			if test.expectRole != arn.Role {
				t.Errorf("expected %q, received %q", test.expectRole, arn.Role)
			}
		})
	}
}
