package main

import (
	"bitbucket.org/emindsys/onelogin-aws-cli/onelogin"
	"fmt"
	"os"
)

func main() {
	var secret string = os.Getenv("ONELOGIN_CLIENT_SECRET")
	var id string = os.Getenv("ONELOGIN_CLIENT_ID")

	response := onelogin.GenerateToken(secret, id)

	fmt.Println("Response: %v", response)
}
