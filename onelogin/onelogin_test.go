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
