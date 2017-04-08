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

func GenerateTokens(secret string, id string) (error, *GenerateTokensResponse) {
	// Construct HTTP request
	var data = []byte(`{"grant_type":"client_credentials"}`)
	req, err := http.NewRequest(
		"POST",
		"https://api.us.onelogin.com/auth/oauth2/token",
		bytes.NewBuffer(data),
	)
	req.Header.Set(
		"Authorization",
		fmt.Sprintf("client_id:%v, client_secret:%v", id, secret),
	)
	req.Header.Set("Content-Type", "application/json")

	// Send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("An error has occurred")
	}

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Request failed: Got %d as a response", resp.StatusCode)), nil
	}

	// Get data from response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	bodyBytes := []byte(body)

	// Parse JSON
	var generateTokenResponse GenerateTokensResponse
	if err := json.Unmarshal(bodyBytes, &generateTokenResponse); err != nil {
		panic(err)
	}

	return nil, &generateTokenResponse
}
