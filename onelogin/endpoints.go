/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package onelogin

import (
	"fmt"
	"net/url"
)

const (
	usBase = "https://api.us.onelogin.com"
	euBase = "https://api.eu.onelogin.com"

	// GenerateSamlAssertionPath - OneLogin API endpoint to generate a SAML assertions
	GenerateSamlAssertionPath string = "/api/2/saml_assertion"

	// GenerateTokensPath - OneLogin API endpoint to generate an access token and refresh token
	GenerateTokensPath string = "/auth/oauth2/v2/token"

	// GetUserByEmailPath - OneLogin API endpoint to get a paginated list of users via email address
	GetUserByEmailPath string = "/api/2/users?email=%s"

	// VerifyFactorPath - OneLogin API endpoint to verify a one-time password (OTP) value
	VerifyFactorPath string = "/api/2/saml_assertion/verify_factor"
)

// Endpoints represent the OneLogin API HTTP endpoints.
type Endpoints struct {
	Region string

	base *url.URL
}

func (e *Endpoints) setBase() (err error) {
	var base string
	switch e.Region {
	case "US":
		base = usBase

	case "EU":
		base = euBase

	default:
		return fmt.Errorf("region %q is an invalid OneLogin region. Valid values are EU or US", e.Region)
	}

	e.base, err = url.Parse(base)

	return
}

// GenerateSamlAssertion will return a the relevant Generate SAML Assertion
// endpoint for a given base URL
func (e Endpoints) GenerateSamlAssertion() string {
	return e.doURL(GenerateSamlAssertionPath, make(url.Values))
}

// GenerateTokens will return the relevant Generate Tokens endpoint
// for a base URL
func (e Endpoints) GenerateTokens() string {
	return e.doURL(GenerateTokensPath, make(url.Values))
}

// GetUserByEmail will, given an email address, return a valid url
// to search the Users endpoint by email address
func (e Endpoints) GetUserByEmail(email string) string {
	return e.doURL(GetUserByEmailPath, url.Values{"email": []string{email}})
}

// VerifyFactor will return a valid URL for requests to check MFA tokens
func (e Endpoints) VerifyFactor() string {
	return e.doURL(VerifyFactorPath, make(url.Values))
}

func (e Endpoints) doURL(endpoint string, params url.Values) string {
	if e.base == nil {
		return ""
	}

	e.base.Path = endpoint
	e.base.RawQuery = params.Encode()

	return e.base.String()
}
