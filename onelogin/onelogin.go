package onelogin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// TODO Review error handling
// TODO Move to named returns?
// TODO Convert prints to logging

// TODO Add support for eu.onelogin.com
const (
	GenerateTokensUrl        string = "https://api.us.onelogin.com/auth/oauth2/token"
	GenerateSamlAssertionUrl string = "https://api.us.onelogin.com/api/1/saml_assertion"
	VerifyFactorUrl          string = "https://api.us.onelogin.com/api/1/saml_assertion/verify_factor"
)

var Client = http.Client{}

type GenerateTokensParams struct {
	GrantType string `json:"grant_type"`
}

type GenerateTokensResponse struct {
	Status struct {
		Error   bool   `json:"error"`
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"status"`
	Data []struct {
		AccessToken  string    `json:"access_token"`
		CreatedAt    time.Time `json:"created_at"`
		ExpiresIn    int       `json:"expires_in"`
		RefreshToken string    `json:"refresh_token"`
		TokenType    string    `json:"token_type"`
		AccountID    int       `json:"account_id"`
	} `json:"data"`
}

type GenerateSamlAssertionParams struct {
	UsernameOrEmail string `json:"username_or_email"`
	Password        string `json:"password"`
	AppId           string `json:"app_id"`
	Subdomain       string `json:"subdomain"`
	IpAddress       string `json:"ip_address"`
}

// TODO This one assumes MFA is enabled. Need to handle all cases.
type GenerateSamlAssertionResponse struct {
	Status struct {
		Error   bool   `json:"error"`
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"status"`
	Data []struct {
		StateToken string `json:"state_token"`
		Devices    []struct {
			DeviceId   int    `json:"device_id"`
			DeviceType string `json:"device_type"`
		}
		CallbackUrl string `json:"callback_url"`
		User        struct {
			Lastname  string `json:"lastname"`
			Username  string `json:"username"`
			Email     string `json:"email"`
			Firstname string `json:"firstname"`
			Id        int    `json:"id"`
		}
	}
}

type VerifyFactorParams struct {
	AppId      string `json:"app_id"`
	DeviceId   string `json:"device_id"`
	StateToken string `json:"state_token"`
	OtpToken   string `json:"otp_token"`
}

type VerifyFactorResponse struct {
	Status struct {
		Error   bool   `json:"error"`
		Code    int    `json:"code"`
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"status"`
	Data string `json:"data"`
}

// OneLoginError represents an error response received from the OneLogin API. In addition
// to the standard error string it contains the HTTP status code to assist in identifying
// the real cause for the error. This is necessary because the actual calls to the OneLogin
// API are made inside doRequest(), so we need a way to determine the HTTP status code outside
// that function.
//type OneLoginError struct {
//	err        string
//	StatusCode int
//}
//
//func (e *OneLoginError) Error() string {
//	return e.err
//}

// createRequest constructs an HTTP request and returns a pointer to it.
// TODO Wrap arguments in a type
func createRequest(method string, url string, headers map[string]string, body interface{}) (*http.Request, error) {
	json, err := json.Marshal(body)
	if err != nil {
		return nil, errors.New("Error parsing body")
	}

	req, err := http.NewRequest(
		method,
		url,
		bytes.NewBuffer(json),
	)
	if err != nil {
		return nil, errors.New("Failed to create HTTP request")
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

// DoRequest gets a pointer to an HTTP request and an HTTP client, executes the request
// using the client, handles any HTTP-related errors and returns any data as a string.
func doRequest(c *http.Client, r *http.Request) (string, error) {
	resp, err := c.Do(r)
	if err != nil {
		return "", errors.New("Could not send HTTP request")
	}

	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	b := []byte(body)

	return string(b), nil
}

// handleResponse gets a JSON-encoded HTTP response data and loads it into the given struct.
func handleResponse(j string, d interface{}) error {
	err := json.Unmarshal([]byte(j), d)
	if err != nil {
		return errors.New("Couldn't parse JSON")
	}

	return nil
}

// GenerateTokens generates the tokens required for interacting with the OneLogin
// API.
func GenerateTokens(clientId, clientSecret string) (string, error) {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("client_id:%v, client_secret:%v", clientId, clientSecret),
		"Content-Type":  "application/json",
	}
	body := GenerateTokensParams{GrantType: "client_credentials"}

	req, err := createRequest(
		http.MethodPost,
		GenerateTokensUrl,
		headers,
		&body,
	)
	if err != nil {
		return "", errors.New("Could not create request")
	}

	data, err := doRequest(&Client, req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %v", err)
	}

	var resp GenerateTokensResponse

	if err := handleResponse(data, &resp); err != nil {
		return "", fmt.Errorf("Could not parse HTTP response: %v", err)
	}

	return resp.Data[0].AccessToken, nil
}

// GenerateSamlAssertion gets a OneLogin access token and a GenerateSamlAssertionParams struct
// and returns a GenerateSamlAssertionResponse.
// TODO improve doc
func GenerateSamlAssertion(token string, p *GenerateSamlAssertionParams) (*GenerateSamlAssertionResponse, error) {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("bearer:%v", token),
		"Content-Type":  "application/json",
	}
	body := p

	req, err := createRequest(
		http.MethodPost,
		GenerateSamlAssertionUrl,
		headers,
		&body,
	)
	if err != nil {
		return nil, errors.New("Could not create request")
	}

	data, err := doRequest(&Client, req)
	if err != nil {
		//if oneLoginError, ok := err.(*OneLoginError); ok {
		//	fmt.Println(oneLoginError.StatusCode)
		//}
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}

	var resp GenerateSamlAssertionResponse

	if err := handleResponse(data, &resp); err != nil {
		return nil, fmt.Errorf("Could not parse HTTP response: %v", err)
	}

	return &resp, nil
}

// VerifyFactor gets a OneLogin access token and a VerifyFactorParams struct and returns a
// VerifyFactorResponse.
func VerifyFactor(token string, p *VerifyFactorParams) (*VerifyFactorResponse, error) {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("bearer:%v", token),
		"Content-Type":  "application/json",
	}
	body := p

	req, err := createRequest(
		http.MethodPost,
		VerifyFactorUrl,
		headers,
		&body,
	)
	if err != nil {
		// TODO Let the user know which method generated the error
		return nil, errors.New("Could not create request")
	}

	data, err := doRequest(&Client, req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}

	var resp VerifyFactorResponse

	if err := handleResponse(data, &resp); err != nil {
		return nil, fmt.Errorf("Could not parse HTTP response: %v", err)
	}

	return &resp, nil
}
