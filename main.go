package main

import (
	"fmt"
	"os"

	"bitbucket.org/emindsys/onelogin-aws-cli/onelogin"
	"github.com/howeyc/gopass"
)

func main() {
	// TODO Add handling for missing env vars
	var secret string = os.Getenv("ONELOGIN_CLIENT_SECRET")
	var id string = os.Getenv("ONELOGIN_CLIENT_ID")
	var appId = os.Args[1]

	// Get OneLogin access token
	err, token := onelogin.GenerateTokens(id, secret)
	if err != nil {
		panic(err)
	}

	// Get credentials from the user
	fmt.Print("OneLogin username: ")
	var user string
	fmt.Scanln(&user)

	fmt.Print("OneLogin password: ")
	pass, err := gopass.GetPasswd()
	if err != nil {
		panic("Couldn't read password from terminal")
	}

	// Generate SAML assertion
	p := onelogin.GenerateSamlAssertionParams{
		UsernameOrEmail: user,
		Password:        string(pass),
		AppId:           appId,
		Subdomain:       "emind",
	}

	err, resp := onelogin.GenerateSamlAssertion(token, &p)
	if err != nil {
		panic(err)
	}

	fmt.Println(resp.Data[0].StateToken)
}
