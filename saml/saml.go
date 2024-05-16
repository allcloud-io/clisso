/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package saml

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/allcloud-io/clisso/log"
	"github.com/crewjam/saml"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type ARN struct {
	Role     string
	Provider string
	Name     string
}

const roleSAMLAttributeName = "https://aws.amazon.com/SAML/Attributes/Role"
const roleRegex = `^arn:(?:aws|aws-cn):iam::(?P<Id>\d+):(?P<Name>role\/\S+)$`
const idpRegex = `^arn:(?:aws|aws-cn):iam::\d+:saml-provider\/\S+$`

func Get(data, pArn string) (a ARN, err error) {
	samlBody, err := decode(data)
	if err != nil {
		log.Log.WithError(err).Error("Error decoding SAML assertion")
		return
	}

	x := new(saml.Response)
	err = xml.Unmarshal(samlBody, x)
	if err != nil {
		log.Log.WithError(err).Error("Error parsing SAML assertion")
		return
	}

	arns := extractArns(x.Assertion.AttributeStatements, pArn)

	switch len(arns) {
	case 0:
		err = errors.New("no valid AWS roles were returned")
		return

	case 1:
		a = arns[0]

		return
	}

	// Multiple ARNs returned - ask user which one to use.
	a = arns[ask(arns)]

	return
}

func decode(in string) (b []byte, err error) {
	return base64.StdEncoding.DecodeString(in)
}

func extractArns(stmts []saml.AttributeStatement, pArn string) []ARN {
	log.Log.WithField("preferredARN", pArn).Trace("Extracting ARNs from SAML AttributeStatements")
	// check for human readable ARN strings in config
	accounts := viper.GetStringMap("global.accounts")
	log.Log.WithFields(accounts).Trace("Accounts loaded from config")
	arns := make([]ARN, 0)

	for _, stmt := range stmts {
		for _, attr := range stmt.Attributes {
			if attr.Name == roleSAMLAttributeName {
				for _, av := range attr.Values {
					// Value is empty
					if len(av.Value) == 0 {
						// could be that other attribute statements will contain data, so we no longer return but continue
						continue
					}

					// Verify we have one of the following formats:
					// 1. arn:aws:iam::xxxxxxxxxxxx:role/MyRole,arn:aws:iam::xxxxxxxxxxxx:saml-provider/MyProvider
					// 2. arn:aws:iam::xxxxxxxxxxxx:saml-provider/MyProvider,arn:aws:iam::xxxxxxxxxxxx:role/MyRole
					// Error otherwise.
					components := strings.Split(strings.TrimSpace(av.Value), ",")
					if len(components) != 2 {
						// Wrong number of components - move on
						log.Log.WithFields(logrus.Fields{
							"components": components,
							"length":     len(components),
							"value":      av.Value,
						}).Trace("Wrong number of components in Value")
						continue
					}

					// people like to put spaces in there, AWS accepts them, let's remove them on our end too.
					components[0] = strings.TrimSpace(components[0])
					components[1] = strings.TrimSpace(components[1])
					log.Log.WithField("components", components).Trace("ARN components extracted from SAML assertion")

					arn := ARN{}

					// Logic here for "preferred arn" for the desired account.
					// If pArn is empty, it proceeds as normal.
					// Otherwise it matches it with what is in the .clisso.yaml file
					if pArn != "" {
						if components[0] == pArn {
							log.Log.Trace("Preferred ARN matches first component")
							arn = ARN{components[0], components[1], ""}
						} else if components[1] == pArn {
							log.Log.Trace("Preferred ARN matches second component")
							arn = ARN{components[1], components[0], ""}
						} else {
							log.Log.Trace("Preferred ARN does not match either component")
							continue
						}
					} else {
						// Prepare patterns
						role := regexp.MustCompile(roleRegex)
						idp := regexp.MustCompile(idpRegex)

						if role.MatchString(components[0]) && idp.MatchString(components[1]) {
							log.Log.Trace("First component is role, second component is IdP")
							arn = ARN{components[0], components[1], ""}
						} else if role.MatchString(components[1]) && idp.MatchString(components[0]) {
							log.Log.Trace("First component is IdP, second component is role")
							arn = ARN{components[1], components[0], ""}
						} else {
							log.Log.Trace("Neither component matches expected pattern")
							continue
						}

						// Look up the human friendly name, if available
						if len(accounts) > 0 {
							ids := role.FindStringSubmatch(arn.Role)
							log.Log.WithField("matches", ids).Trace("Role regex matches")

							// if the regex matches we should have 3 entries from the regex match
							// 1) the matching string
							// 2) the match for Id
							// 3) the match for Name
							// we want to match the Id to any accounts/roles in our config
							if len(ids) == 3 && accounts[ids[1]] != "" && accounts[ids[1]] != nil {
								log.Log.Trace("Found human friendly name for account")
								arn.Name = fmt.Sprintf("%s - %s", accounts[ids[1]].(string), ids[2])
							}
						}
					}

					arns = append(arns, arn)
				}

				return arns
			}
		}
	}
	log.Log.Trace("No statements in SAML assertion or no ARNs found.")
	// Empty :(
	return arns
}

func ask(arns []ARN) (idx int) {
	for {
		for i, a := range arns {
			name := a.Role
			// Add the human friendly name if available
			if a.Name != "" {
				name = a.Name
			}

			// Use one-based indexing for human-friendliness.
			fmt.Printf("%d. %s\n", i+1, name)
		}

		var input string
		fmt.Print("Please select an IAM role to assume: ")
		_, err := fmt.Scanln(&input)
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		// Verify we got an integer.
		selected, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Invalid input '%s'\n", input)
			continue
		}

		// Verify selection is within range.
		if selected < 1 || selected > len(arns) {
			fmt.Printf("Invalid value %d. Valid values: 1-%d\n", selected, len(arns))
			continue
		}

		// Translate user-selected index back to zero-based index.
		idx = selected - 1
		break
	}

	return
}
