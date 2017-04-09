package main

import (
	"bitbucket.org/emindsys/onelogin-aws-cli/onelogin"
	"fmt"
	"os"
	"bufio"
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

func getSaml(token, user, pass, appId, ipAddress, subdomain string) string {
	fmt.Print("Generating SAML assertion... ")
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
		fmt.Println("Couldn't get SAML assertion: ", err)
		os.Exit(2)
	}

	return resp.Data[0].StateToken
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

	st := getSaml(t, user, pass, "123456", "testing", "")
	fmt.Println("State token: ", st)
}
