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

// GetSessionTokenResponse represents the result of GetSessionToken.
type GetSessionTokenResponse struct {
	ExpiresAt    time.Time
	Status       string
	SessionToken string
}

// GetSessionToken performs a login operation against the Okta API and returns a session token upon
// successful login.
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
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
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
