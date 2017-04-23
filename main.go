package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

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
	log.Println("Generating OneLogin access tokens")
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
	log.Println("Generating SAML assertion")
	pSaml := onelogin.GenerateSamlAssertionParams{
		UsernameOrEmail: user,
		Password:        string(pass),
		AppId:           appId,
		Subdomain:       "emind",
	}

	rSaml, err := onelogin.GenerateSamlAssertion(token, &pSaml)
	if err != nil {
		log.Fatal(err)
	}

	st := rSaml.Data[0].StateToken
	// TODO Handle multiple devices
	deviceId := strconv.Itoa(rSaml.Data[0].Devices[0].DeviceId)

	fmt.Print("Please enter your OneLogin OTP: ")
	var otp string
	fmt.Scanln(&otp)

	// Verify MFA
	pMfa := onelogin.VerifyFactorParams{
		AppId:      appId,
		DeviceId:   string(deviceId),
		StateToken: st,
		OtpToken:   otp,
	}

	rMfa, err := onelogin.VerifyFactor(token, &pMfa)
	if err != nil {
		log.Fatal(err)
	}

	samlAssertion := rMfa.Data
	log.Println(samlAssertion)
}
