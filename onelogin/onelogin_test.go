package onelogin

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

func TestGenerateTokens(t *testing.T) {
	data := `{
    "status": {
        "error": false,
        "code": 200,
        "type": "success",
        "message": "Success"
    },
    "data": [
        {
            "access_token": "fake_token",
            "created_at": "2015-11-11T03:36:18.714Z",
            "expires_in": 36000,
            "refresh_token": "fake",
            "token_type": "bearer",
            "account_id": 555555
        }
 ]
}`

	ts := getTestServer(data)
	defer ts.Close()

	resp, err := GenerateTokens(ts.URL, "test", "test")
	if err != nil {
		t.Errorf("GenerateTokens failed: %s", err)
	}
	if resp != "fake_token" {
		t.Errorf("Wrong response, got: %v, want: %v", resp, "fake_token")
	}
}

func TestGenerateSamlAssertion(t *testing.T) {
	data := `{
    "status": {
        "type": "success",
        "message": "MFA is required for this user",
        "code": 200,
        "error": false
    },
    "data": [
        {
            "state_token": "fake_state_token",
            "devices": [
                {
                    "device_id": 666666,
                    "device_type": "Google Authenticator"
                }
            ],
            "callback_url": "https://api.us.onelogin.com/api/1/saml_assertion/verify_factor",
            "user": {
                "lastname": "test",
                "username": "test",
                "email": "test@onelogin.com",
                "firstname": "test",
                "id": 88888888
            }
        }
    ]
}`
	ts := getTestServer(data)
	defer ts.Close()

	p := GenerateSamlAssertionParams{
		UsernameOrEmail: "test",
		Password:        "test",
		AppId:           "test",
		Subdomain:       "test",
	}

	resp, err := GenerateSamlAssertion(ts.URL, "test", &p)
	if err != nil {
		t.Errorf("GenerateSamlAssertion: %s", err)
	}
	if resp.Data[0].StateToken != "fake_state_token" {
		t.Errorf(
			"Wrong response, got: %v, want: %v",
			resp.Data[0].StateToken, "fake_state_token",
		)
	}
}
