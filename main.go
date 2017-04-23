package main

import (
	"fmt"
	"log"
	"os"

	"bitbucket.org/emindsys/onelogin-aws-cli/onelogin"
	"github.com/howeyc/gopass"
)

func main() {
	// TODO Use Cobra?

	// Get CLI arguments
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %v <app_id>\n", os.Args[0])
		os.Exit(1)
	}
	var appId = os.Args[1]

	// Get env vars
	var secret string = os.Getenv("ONELOGIN_CLIENT_SECRET")
	var id string = os.Getenv("ONELOGIN_CLIENT_ID")

	if secret == "" {
		log.Fatal("The ONELOGIN_CLIENT_SECRET environment variable must bet set.")
	}

	if id == "" {
		log.Fatal("The ONELOGIN_CLIENT_ID environment variable must bet set.")
	}

	// Get OneLogin access token
	token, err := onelogin.GenerateTokens(id, secret)
	if err != nil {
		log.Fatal(err)
	}

	// Get credentials from the user
	fmt.Print("OneLogin username: ")
	var user string
	fmt.Scanln(&user)

	fmt.Print("OneLogin password: ")
	pass, err := gopass.GetPasswd()
	if err != nil {
		log.Fatal("Couldn't read password from terminal")
	}

	// Generate SAML assertion
	p := onelogin.GenerateSamlAssertionParams{
		UsernameOrEmail: user,
		Password:        string(pass),
		AppId:           appId,
		Subdomain:       "emind",
	}

	resp, err := onelogin.GenerateSamlAssertion(token, &p)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(resp.Data[0].StateToken)
}
