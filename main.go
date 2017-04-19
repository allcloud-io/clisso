package main

import (
	"fmt"
	"net/http"
	"os"

	"bitbucket.org/emindsys/onelogin-aws-cli/onelogin"
	//"github.com/howeyc/gopass"
	"encoding/json"
)

func main() {
	// TODO Add handling for missing env vars
	var secret string = os.Getenv("ONELOGIN_CLIENT_SECRET")
	var id string = os.Getenv("ONELOGIN_CLIENT_ID")
	//var appId = os.Args[1]

	// Create HTTP client
	// TODO find a good way to pass the client to the functions in onelogin.go
	c := http.Client{}

	// Get OneLogin access token
	headers := map[string]string{
		"Authorization": fmt.Sprintf("client_id:%v, client_secret:%v", id, secret),
		"Content-Type": "application/json",
	}
	params := onelogin.GenerateTokensParams{
		GrantType: "client_credentials",
	}

	err, req := onelogin.CreateRequest(
		http.MethodPost,
		onelogin.GenerateTokensUrl,
		headers,
		&params,
	)
	if err != nil {
		panic(err)
	}

	err, data := onelogin.DoRequest(&c, req)
	if err != nil {
		panic(err)
	}

	var resp onelogin.GenerateTokensResponse

	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		panic(err)
	}

	fmt.Println(resp.Data[0].AccessToken)
}
