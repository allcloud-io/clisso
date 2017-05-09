package onelogin

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func getTestServer(resp string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(resp))
	}))

	return ts
}

func TestGenerateTokens(t *testing.T) {
	resp := `{
    "status": {
        "error": false,
        "code": 200,
        "type": "success",
        "message": "Success"
    },
    "data": [
        {
            "access_token": "fake",
            "created_at": "2015-11-11T03:36:18.714Z",
            "expires_in": 36000,
            "refresh_token": "fake",
            "token_type": "bearer",
            "account_id": 555555
        }
 ]
}`

	ts := getTestServer(resp)
	defer ts.Close()

	_, err := GenerateTokens(ts.URL, "test", "test")
	if err != nil {
		t.Errorf("Oops: %s", err)
	}
}

func TestGenerateSamlAssertion(t *testing.T) {
	resp := `{
    "status": {
        "type": "success",
        "message": "MFA is required for this user",
        "code": 200,
        "error": false
    },
    "data": [
        {
            "state_token": "fake",
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
	ts := getTestServer(resp)
	defer ts.Close()

	p := GenerateSamlAssertionParams{
		UsernameOrEmail: "test",
		Password:        "test",
		AppId:           "test",
		Subdomain:       "test",
	}

	_, err := GenerateSamlAssertion(ts.URL, "test", &p)
	if err != nil {
		t.Errorf("Oops: %s", err)
	}
}
