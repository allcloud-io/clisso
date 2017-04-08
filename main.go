package main

import (
	"bitbucket.org/emindsys/onelogin-aws-cli/onelogin"
	"fmt"
	"os"
)

func main() {
	// TODO Add handling for missing env vars
	var secret string = os.Getenv("ONELOGIN_CLIENT_SECRET")
	var id string = os.Getenv("ONELOGIN_CLIENT_ID")

	err, response := onelogin.GenerateTokens(secret, id)

	if err != nil {
		fmt.Println("Token generation failed: ", err)
	} else {
		fmt.Printf("Access token: %v\n", response.Data[0].AccessToken)
	}
}
