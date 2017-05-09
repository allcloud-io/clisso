package onelogin

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type FakeHTTPClient struct{}

func (c *FakeHTTPClient) Do(r *http.Request) (*http.Response, error) {
	fmt.Println("Fake HTTP client called")

	resp, err := c.Do(r)
	if err != nil {
		return nil, errors.New("Error processing HTTP request")
	}

	return resp, nil
}

func TestGenerateTokens(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
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
}`))
	}))
	defer ts.Close()

	//	Client = &FakeHTTPClient{}
	_, err := GenerateTokens(ts.URL, "abcd", "efgh")
	if err != nil {
		t.Errorf("Oops: %s", err)
	}
}
