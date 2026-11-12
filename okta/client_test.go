/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package okta

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/allcloud-io/clisso/log"
)

var _, _ = log.SetupLogger("panic", "", false, true)

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

func TestGetSessionToken(t *testing.T) {
	data := `{
		"expiresAt": "2018-11-03T10:15:57.000Z",
		"status": "SUCCESS",
		"sessionToken": "fake_token"
	}`

	ts := getTestServer(data)
	defer ts.Close()

	c.BaseURL = ts.URL

	resp, err := c.GetSessionToken(&GetSessionTokenParams{Username: "test", Password: "test"})
	if err != nil {
		t.Errorf("getting session token: %s", err)
	}
	if resp.SessionToken != "fake_token" {
		t.Errorf("Wrong response, got: %v, want: %v", resp.SessionToken, "fake_token")
	}
}

func TestGetSessionTokenMFARequired(t *testing.T) {
	data := `{
		"stateToken": "fake_token",
		"expiresAt": "2018-07-08T13:47:53.000Z",
		"status": "MFA_REQUIRED",
		"_embedded": {
			"factors": [
				{
					"id": "fake_id",
					"factorType": "token:software:totp",
					"provider": "GOOGLE",
					"vendorName": "GOOGLE",
					"profile": {
						"credentialId": "test@test.local"
					},
					"_links": {
						"verify": {
							"href": "https://test.test.local/api/v1/authn/factors/fake/verify",
							"hints": {
								"allow": [
									"POST"
								]
							}
						}
					}
				}
			]
		}
	}`

	ts := getTestServer(data)
	defer ts.Close()

	c.BaseURL = ts.URL

	resp, err := c.GetSessionToken(&GetSessionTokenParams{Username: "test", Password: "test"})
	if err != nil {
		t.Errorf("getting session token: %v", err)
	}
	if resp.StateToken != "fake_token" {
		t.Errorf("Wrong response, got: %v, want: %v", resp.StateToken, "fake_token")
	}
}

func TestVerifyFactor(t *testing.T) {
	data := `{
		"expiresAt": "2015-11-03T10:15:57.000Z",
		"status": "SUCCESS",
		"sessionToken": "fake_token"
	}`

	ts := getTestServer(data)
	defer ts.Close()

	c.BaseURL = ts.URL

	resp, err := c.VerifyFactor(&VerifyFactorParams{
		FactorID:   "fake_id",
		StateToken: "fake_state_token",
		PassCode:   "123456",
	})
	if err != nil {
		t.Errorf("verifying factor: %v", err)
	}

	if resp.Status != "SUCCESS" {
		t.Errorf("Wrong response, got: %v, want: %v", resp.Status, "SUCCESS")
	}
	if resp.SessionToken != "fake_token" {
		t.Errorf("Wrong response, got: %v, want: %v", resp.SessionToken, "fake_token")
	}
	exp, _ := time.Parse("2006-01-02T15:04:05.000Z", "2015-11-03T10:15:57.000Z")
	if resp.ExpiresAt != exp {
		t.Errorf("Wrong response, got: %v, want: %v", resp.ExpiresAt, exp)
	}
}

func TestLaunchApp(t *testing.T) {
	for _, tc := range []struct {
		name      string
		status    int
		body      string
		want      string
		expectErr bool
	}{
		{"extracts SAMLResponse", 200,
			`<html><body><form id="appForm" action="x" method="POST"><input name="SAMLResponse" type="hidden" value="PHNhbWw+YWJjPC9zYW1sPg=="/><input name="RelayState" value="r"/></form></body></html>`,
			"PHNhbWw+YWJjPC9zYW1sPg==", false},
		{"no form", 200, `<html><body><p>nothing</p></body></html>`, "", false},
		{"http error", 500, `oops`, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("sessionToken") != "tok" {
					t.Errorf("sessionToken not passed, got %q", r.URL.RawQuery)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer ts.Close()

			cl, err := NewClient(ts.URL)
			if err != nil {
				t.Fatal(err)
			}
			got, err := cl.LaunchApp(&LaunchAppParams{SessionToken: "tok", URL: ts.URL})
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if *got != tc.want {
				t.Errorf("got %q, want %q", *got, tc.want)
			}
		})
	}
}
