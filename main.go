package main

import (
	"fmt"
	"os"

	"bitbucket.org/emindsys/onelogin-aws-cli/onelogin"
	//"github.com/howeyc/gopass"
)

func main() {
	// TODO Add handling for missing env vars
	var secret string = os.Getenv("ONELOGIN_CLIENT_SECRET")
	var id string = os.Getenv("ONELOGIN_CLIENT_ID")
	//var appId = os.Args[1]

	// Get OneLogin access token
	err, token := onelogin.GenerateTokens(id, secret)
	if err != nil {
		panic(err)
	}

	fmt.Println(token)
}
