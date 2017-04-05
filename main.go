package main

import (
	"bitbucket.org/emindsys/onelogin-aws-cli/onelogin"
	"fmt"
	"os"
)

func main() {
	var secret string = os.Getenv("ONELOGIN_CLIENT_SECRET")
	var id string = os.Getenv("ONELOGIN_CLIENT_ID")

	err, response := onelogin.GenerateTokens(secret, id)

	if err != nil {
		fmt.Println("Token generation failed: ", err)
	} else {
		fmt.Printf("Response: %v", response)
	}
}
