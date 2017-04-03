package onelogin

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"bytes"
)

func GenerateToken(secret string, id string) (string) {
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

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("An error has occurred")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	return string(body)
}
