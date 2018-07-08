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
		t.Errorf("GenerateTokens failed: %s", err)
	}
	if resp.SessionToken != "fake_token" {
		t.Errorf("Wrong response, got: %v, want: %v", resp, "fake_token")
	}
}
