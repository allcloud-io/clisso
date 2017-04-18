package main

import (
	"bitbucket.org/emindsys/onelogin-aws-cli/onelogin"
	"fmt"
	"os"
	"bufio"
	"errors"
)

func getToken(secret, id string) string {
	fmt.Print("Generating OneLogin API token... ")
	p := onelogin.GenerateTokensParams{Secret: secret, Id: id}
	err, resp := onelogin.GenerateTokens(&p)
	fmt.Println("done")

	if err != nil {
		fmt.Println("Token generation failed: ", err)
		os.Exit(2)
	}
	return resp.Data[0].AccessToken
}

func getSaml(token, user, pass, appId, ipAddress, subdomain string) (error, string) {
	fmt.Print("Requesting SAML assertion... ")
	pSaml := onelogin.GenerateSamlAssertionParams{}
	pSaml.Headers.AccessToken = token
	pSaml.RequestData.UsernameOrEmail = user
	pSaml.RequestData.Password = pass
	pSaml.RequestData.AppId = appId
	pSaml.RequestData.Subdomain = subdomain
	pSaml.RequestData.IpAddress = ipAddress
	err, resp := onelogin.GenerateSamlAssertion(&pSaml)
	fmt.Println("done")

	if err != nil {
		fmt.Println("Couldn't get SAML assertion")
		fmt.Println(err)
		os.Exit(2)
	}

	// Handle response
	status := resp.Status

	if status.Type != "success" {
		return errors.New(fmt.Sprintf("SAML assertion failed: %v", status.Message)), ""
	}

	data := resp.Data[0]

	return nil, data.StateToken
}

func main() {
	// TODO Add handling for missing env vars
	var secret string = os.Getenv("ONELOGIN_CLIENT_SECRET")
	var id string = os.Getenv("ONELOGIN_CLIENT_ID")

	// Get OneLogin access token
	t := getToken(secret, id)

	// Get credentials from user
	r := bufio.NewReader(os.Stdin)
	fmt.Print("OneLogin username: ")
	user, _ := r.ReadString('\n')
	fmt.Print("OneLogin password: ")
	pass, _ := r.ReadString('\n')

	err, st := getSaml(t, user, string(pass), appId, "", "emind")
	if err != nil {
		fmt.Println("Couldn't get state token")
		fmt.Println(err)
		os.Exit(2)
	}
	fmt.Println("State token: ", st)
}
