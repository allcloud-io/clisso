package okta

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func getTestServer(data string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(data))
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
