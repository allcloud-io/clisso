package okta

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	StatusSuccess     = "SUCCESS"
	StatusMFARequired = "MFA_REQUIRED"
)

// Client represents an Okta API client.
type Client struct {
	http.Client
	BaseURL string
}

// GetSessionTokenParams represents the parameters for GetSessionToken.
type GetSessionTokenParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// GetSessionTokenResponse represents the result of a call to GetSessionToken.
type GetSessionTokenResponse struct {
	ExpiresAt    time.Time `json:"expiresAt"`
	SessionToken string    `json:"sessionToken"`
	StateToken   string    `json:"stateToken"`
	Status       string    `json:"status"`
	Embedded     struct {
		Factors []struct {
			ID    string `json:"id"`
			Links struct {
				Verify struct {
					Href string `json:"href"`
				} `json:"verify"`
			} `json:"_links"`
		} `json:"factors"`
	} `json:"_embedded"`
}

// GetSessionToken performs a login operation against the Okta API and returns a session token upon
// successful login.
//
// Following a successful call (error == nil), the Status field of the response must be checked. If
// the status is StatusSuccess then the SessionToken field contains a valid session token and the
// authentication action is complete. If the status is StatusMFARequired, the user needs to provide
// an MFA one-time password before a session token can be retrieved. In this case, the StateToken
// field will contain the state token to pass to the MFA verification API endpoint, and the
// Embedded field will contain information about the available factor(s). The caller will then need
// to call the VerifyFactor function to complete the authentication and obtain a session token.
// See the Okta API documentation for more details:
// https://developer.okta.com/docs/api/resources/authn#verify-totp-factor
func (c *Client) GetSessionToken(p *GetSessionTokenParams) (*GetSessionTokenResponse, error) {
	h := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}
	req, err := makeRequest(http.MethodPost, c.BaseURL+"/api/v1/authn", h, p)
	if err != nil {
		return nil, fmt.Errorf("creating request: %v", err)
	}

	data, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("doing HTTP request: %v", err)
	}

	var resp GetSessionTokenResponse
	err = json.Unmarshal([]byte(data), &resp)
	if err != nil {
		return nil, fmt.Errorf("parsing HTTP response: %v", err)
	}

	return &resp, nil
}

// VerifyFactorParams represents the parameters for VerifyFactor.
type VerifyFactorParams struct {
	FactorID   string `json:"factorId"`
	StateToken string `json:"stateToken"`
	PassCode   string `json:"passCode"`
}

// VerifyFactorResponse represents the result of a call to VerifyFactor.
type VerifyFactorResponse struct {
	ExpiresAt    time.Time `json:"expiresAt"`
	SessionToken string    `json:"sessionToken"`
	Status       string    `json:"status"`
}

// VerifyFactor performs MFA verification.
func (c *Client) VerifyFactor(p *VerifyFactorParams) (*VerifyFactorResponse, error) {
	h := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}
	url := fmt.Sprintf("%s/api/v1/authn/factors/%s/verify", c.BaseURL, p.FactorID)
	req, err := makeRequest(http.MethodPost, url, h, p)
	if err != nil {
		return nil, fmt.Errorf("creating request: %v", err)
	}

	data, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("doing HTTP request: %v", err)
	}

	var resp VerifyFactorResponse
	err = json.Unmarshal([]byte(data), &resp)
	if err != nil {
		return nil, fmt.Errorf("parsing HTTP response: %v", err)
	}

	return &resp, nil
}

// doRequest gets a pointer to an HTTP request and an HTTP client, executes the request
// using the client, handles any HTTP-related errors and returns any data as a string.
func (c *Client) doRequest(r *http.Request) (string, error) {
	resp, err := c.Do(r)
	if err != nil {
		return "", fmt.Errorf("sending HTTP request: %v", err)
	}

	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	b := []byte(body)

	return string(b), nil
}

// NewClient creates a new Client and returns a pointer to it.
func NewClient(url string) *Client {
	return &Client{BaseURL: url}
}

// makeRequest constructs an HTTP request and returns a pointer to it.
// TODO Wrap arguments in a type
func makeRequest(method string, url string, headers map[string]string, body interface{}) (*http.Request, error) {
	json, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("parsing body: %v", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(json))
	if err != nil {
		return nil, fmt.Errorf("making HTTP request: %v", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req, nil
}
