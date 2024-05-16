/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package onelogin

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/allcloud-io/clisso/log"
)

func getTestServer(data string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(data))
		if err != nil {
			panic(err)
		}
	}))

	return ts
}

var c = Client{}

var _ = log.NewLogger("panic", "", false)

func TestNewClient(t *testing.T) {
	for _, test := range []struct {
		name        string
		region      string
		expectError bool
	}{
		{"Valid region", "US", false},
		{"Invalid region", "invalid", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewClient(test.region)
			if test.expectError && err == nil {
				t.Errorf("expected error")
			}
			if !test.expectError && err != nil {
				t.Errorf("unexpected error %+v", err)
			}
		})
	}
}

func TestGenerateTokens(t *testing.T) {
	data := `{
	"access_token": "fake_token",
	"created_at": "2015-11-11T03:36:18.714Z",
	"expires_in": 36000,
	"refresh_token": "fake",
	"token_type": "bearer",
	"account_id": 555555
}`

	ts := getTestServer(data)
	defer ts.Close()

	c.Endpoints.base, _ = url.Parse(ts.URL)

	resp, err := c.GenerateTokens("test", "test")
	if err != nil {
		t.Errorf("GenerateTokens failed: %s", err)
	}
	if resp != "fake_token" {
		t.Errorf("Wrong response, got: %v, want: %v", resp, "fake_token")
	}
}

func TestGenerateSamlAssertion(t *testing.T) {
	data := `{
	"state_token": "fake_state_token",
	"message": "fake_message",
	"devices": [
		{
			"device_id": 666666,
			"device_type": "Google Authenticator"
		}
	],
	"callback_url": "https://api.us.onelogin.com/api/2/saml_assertion/verify_factor",
	"user": {
		"lastname": "test",
		"username": "test",
		"email": "test@onelogin.com",
		"firstname": "test",
		"id": 88888888
	}
}`
	ts := getTestServer(data)
	defer ts.Close()

	c.Endpoints.base, _ = url.Parse(ts.URL)

	p := GenerateSamlAssertionParams{
		UsernameOrEmail: "test",
		Password:        "test",
		AppId:           "test",
		Subdomain:       "test",
	}

	resp, err := c.GenerateSamlAssertion("test", &p)
	if err != nil {
		t.Errorf("GenerateSamlAssertion: %s", err)
	}
	if resp.StateToken != "fake_state_token" {
		t.Errorf(
			"Wrong response, got: %v, want: %v",
			resp.StateToken, "fake_state_token",
		)
	}
}

func TestVerifyFactor(t *testing.T) {
	data := `{
    "status": {
        "type": "success",
        "message": "Success",
        "code": 200,
        "error": false
    },
    "data": "abcd"
}`

	ts := getTestServer(data)
	defer ts.Close()

	c.Endpoints.base, _ = url.Parse(ts.URL)

	p := VerifyFactorParams{
		AppId:      "test",
		DeviceId:   "test",
		StateToken: "test",
		OtpToken:   "test",
	}

	resp, err := c.VerifyFactor("test", &p)
	if err != nil {
		t.Errorf("VerifyFactor: %s", err)
	}

	if resp.Data != "abcd" {
		t.Errorf(
			"Wrong response, got: %v, want: %v",
			resp.Data, "abcd",
		)
	}
}
