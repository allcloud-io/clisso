package onelogin

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

// TODO Add support for eu.onelogin.com
const (
	GenerateTokensUrl string = "https://api.us.onelogin.com/auth/oauth2/token"
	GenerateSamlAssertionUrl string = "https://api.us.onelogin.com/api/1/saml_assertion"
	VerifyFactorUrl string = "https://api.us.onelogin.com/api/1/saml_assertion/verify_factor"
)

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
	Headers struct {
		AccessToken string
	}
	RequestData struct {
		UsernameOrEmail string `json:"username_or_email"`
		Password        string `json:"password"`
		AppId           string `json:"app_id"`
		Subdomain       string `json:"subdomain"`
		IpAddress       string `json:"ip_address"`
	}
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
	Headers struct {
		AccessToken string
	}
	RequestData struct {
		AppId      string `json:"app_id"`
		DeviceId   string `json:"device_id"`
		StateToken string `json:"state_token"`
		OtpToken   string `json:"otp_token"`
	}
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

// Request constructs an HTTP request and returns a pointer to it.
func CreateRequest(method string, url string, headers map[string]string, data interface{}) (error, *http.Request) {
	// TODO error handling
	json, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest(
		method,
		url,
		bytes.NewBuffer(json),
	)
	if err != nil {
		panic(err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return nil, req
}

// DoRequest gets a pointer to an HTTP request and an HTTP client, executes the request
// using the client, handles any HTTP-related errors and returns any data as a string.
func DoRequest(c *http.Client, r *http.Request) (error, string) {
	resp, err := c.Do(r)
	if err != nil {
		return errors.New("Could not send HTTP request"), ""
	}

	// TODO handle errors

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	b := []byte(body)

	return nil, string(b)
}

// GenerateTokens generates the tokens required for interacting with the OneLogin
// API.
//func GenerateTokens(clientId, clientSecret string) (error, string) {
//	// TODO
//}
