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

type GenerateTokensParams struct {
	Secret string
	Id     string
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

// GenerateTokens generates the tokens required for interacting with the OneLogin
// API.
func GenerateTokens(p *GenerateTokensParams) (error, *GenerateTokensResponse) {
	// Construct HTTP request
	var data = []byte(`{"grant_type":"client_credentials"}`)
	req, err := http.NewRequest(
		"POST",
		"https://api.us.onelogin.com/auth/oauth2/token",
		bytes.NewBuffer(data),
	)
	req.Header.Set(
		"Authorization",
		fmt.Sprintf("client_id:%v, client_secret:%v", p.Id, p.Secret),
	)
	req.Header.Set("Content-Type", "application/json")

	// Send HTTP request
	c := &http.Client{}
	resp, err := c.Do(req)

	if err != nil {
		return errors.New("HTTP request failed"), nil
	}

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("Request failed: Got %d as a response", resp.StatusCode)
		return errors.New(s), nil
	}

	// Get data from response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	b := []byte(body)

	// Parse JSON
	var r GenerateTokensResponse
	if err := json.Unmarshal(b, &r); err != nil {
		panic(err)
	}

	return nil, &r
}

func GenerateSamlAssertion(p *GenerateSamlAssertionParams) (error, *GenerateSamlAssertionResponse) {
	// Construct HTTP request
	j, err := json.Marshal(p.RequestData)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(
		"POST",
		"https://api.us.onelogin.com/api/1/saml_assertion",
		bytes.NewBuffer(j),
	)
	req.Header.Set(
		"Authorization",
		fmt.Sprintf("bearer:%v", p.Headers.AccessToken),
	)
	req.Header.Set("Content-Type", "application/json")

	// Send HTTP request
	c := http.Client{}
	resp, err := c.Do(req)

	if err != nil {
		return errors.New("HTTP request failed"), nil
	}

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("Request failed: Got %d as a response", resp.StatusCode)
		return errors.New(s), nil
	}

	// Get data from response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	b := []byte(body)

	// Parse JSON
	var r GenerateSamlAssertionResponse
	if err := json.Unmarshal(b, &r); err != nil {
		panic(err)
	}

	return nil, &r
}
