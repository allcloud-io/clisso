/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package saml

import (
	"os"
	"testing"

	"github.com/allcloud-io/clisso/log"
	"github.com/crewjam/saml"
)

var _ = log.NewLogger("panic", "", false)

func TestExtractArns(t *testing.T) {
	for _, test := range []struct {
		name      string
		attrs     []saml.Attribute
		expectARN []ARN
	}{
		{"aws-normal", []saml.Attribute{
			{FriendlyName: "Role",
				Name:       "https://aws.amazon.com/SAML/Attributes/Role",
				NameFormat: "string",
				Values:     []saml.AttributeValue{{Value: "arn:aws:iam::1234567890:role/MyRole,arn:aws:iam::1234567890:saml-provider/MyProvider"}}},
		},
			[]ARN{{"arn:aws:iam::1234567890:role/MyRole", "arn:aws:iam::1234567890:saml-provider/MyProvider", ""}}},
		{"aws-reversed", []saml.Attribute{
			{FriendlyName: "Role",
				Name:       "https://aws.amazon.com/SAML/Attributes/Role",
				NameFormat: "string",
				Values:     []saml.AttributeValue{{Value: "arn:aws:iam::1234567890:saml-provider/MyProvider,arn:aws:iam::1234567890:role/MyRole"}}},
		},
			[]ARN{{"arn:aws:iam::1234567890:role/MyRole", "arn:aws:iam::1234567890:saml-provider/MyProvider", ""}}},
		{"aws-spaced-front", []saml.Attribute{
			{FriendlyName: "Role",
				Name:       "https://aws.amazon.com/SAML/Attributes/Role",
				NameFormat: "string",
				Values:     []saml.AttributeValue{{Value: " arn:aws:iam::1234567890:saml-provider/MyProvider,arn:aws:iam::1234567890:role/MyRole"}}},
		},
			[]ARN{{"arn:aws:iam::1234567890:role/MyRole", "arn:aws:iam::1234567890:saml-provider/MyProvider", ""}}},
		{"aws-spaced-end", []saml.Attribute{
			{FriendlyName: "Role",
				Name:       "https://aws.amazon.com/SAML/Attributes/Role",
				NameFormat: "string",
				Values:     []saml.AttributeValue{{Value: "arn:aws:iam::1234567890:saml-provider/MyProvider,arn:aws:iam::1234567890:role/MyRole "}}},
		},
			[]ARN{{"arn:aws:iam::1234567890:role/MyRole", "arn:aws:iam::1234567890:saml-provider/MyProvider", ""}}},
		{"aws-spaced-between", []saml.Attribute{
			{FriendlyName: "Role",
				Name:       "https://aws.amazon.com/SAML/Attributes/Role",
				NameFormat: "string",
				Values:     []saml.AttributeValue{{Value: "arn:aws:iam::1234567890:saml-provider/MyProvider, arn:aws:iam::1234567890:role/MyRole"}}},
		},
			[]ARN{{"arn:aws:iam::1234567890:role/MyRole", "arn:aws:iam::1234567890:saml-provider/MyProvider", ""}}},
		{"aws-cn", []saml.Attribute{
			{FriendlyName: "Role",
				Name:       "https://aws.amazon.com/SAML/Attributes/Role",
				NameFormat: "string",
				Values:     []saml.AttributeValue{{Value: "arn:aws-cn:iam::1234567890:saml-provider/MyProvider,arn:aws-cn:iam::1234567890:role/MyRole"}}},
		},
			[]ARN{{"arn:aws-cn:iam::1234567890:role/MyRole", "arn:aws-cn:iam::1234567890:saml-provider/MyProvider", ""}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			arn := extractArns([]saml.AttributeStatement{{Attributes: test.attrs}}, "")

			if len(test.expectARN) > 0 && len(arn) == 0 {
				t.Fatalf("expected %d arns, received nothing", len(test.expectARN))
			}
			if len(test.expectARN) != len(arn) {
				t.Errorf("expected %d arns, received %d arns", len(test.expectARN), len(arn))
			}
			for i := 0; i < len(test.expectARN); i++ {
				if test.expectARN[i].Name != arn[i].Name {
					t.Errorf("expected %q, received %q", test.expectARN[i].Name, arn[i].Name)
				}

				if test.expectARN[i].Provider != arn[i].Provider {
					t.Errorf("expected %q, received %q", test.expectARN[i].Provider, arn[i].Provider)
				}

				if test.expectARN[i].Role != arn[i].Role {
					t.Errorf("expected %q, received %q", test.expectARN[i].Role, arn[i].Role)
				}
			}
		})
	}
}

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
			b, _ := os.ReadFile(test.path)

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
		name           string
		path           string
		expectProvider string
		expectRole     string
		expectError    bool
	}{
		{
			"Single ARN",
			"testdata/single-arn-response",
			"arn:aws:iam::123456789012:saml-provider/OneLogin-MyProvider",
			"arn:aws:iam::123456789012:role/OneLogin-MyRole",
			false,
		},
		//{"Many ARNs", "testdata/valid-response", "", "", false},         // will ask questions
		{"No ARNs", "testdata/no-arns-resonse", "", "", true},
		{"No ARN value", "testdata/no-arn-value-response", "", "", true},
		{
			"IdP ARN before role ARN",
			"testdata/idp-before-role",
			"arn:aws:iam::123456789012:saml-provider/OneLogin-MyProvider",
			"arn:aws:iam::123456789012:role/OneLogin-MyRole",
			false,
		},
		{"Too many ARN components", "testdata/too-many-components", "", "", true},
		{"Malformed ARN components", "testdata/malformed-components", "", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			b, _ := os.ReadFile(test.path)

			arn, err := Get(string(b), "")
			if test.expectError && err == nil {
				t.Errorf("expected error")
			}
			if !test.expectError && err != nil {
				t.Errorf("unexpected error %+v", err)
			}

			if test.expectProvider != arn.Provider {
				t.Errorf("expected %q, received %q", test.expectProvider, arn.Provider)
			}
			if test.expectRole != arn.Role {
				t.Errorf("expected %q, received %q", test.expectRole, arn.Role)
			}
		})
	}
}
