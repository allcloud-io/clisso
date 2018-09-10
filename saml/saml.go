package saml

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/edaniels/go-saml"
)

type ARN struct {
	Role     string
	Provider string
}

func Get(data string) (a ARN, err error) {
	samlBody, err := decode(data)
	if err != nil {
		return
	}

	x := new(saml.Response)
	err = xml.Unmarshal(samlBody, x)
	if err != nil {
		return
	}

	arns := extractArns(x.Assertion.AttributeStatement.Attributes)

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

func extractArns(attrs []saml.Attribute) (arns []ARN) {
	arns = make([]ARN, 0)

	for _, attr := range attrs {
		if attr.Name == "https://aws.amazon.com/SAML/Attributes/Role" {
			for _, av := range attr.Values {
				// Value is empty
				if len(av.Value) == 0 {
					return
				}
				components := strings.Split(av.Value, ",")

				arns = append(arns, ARN{components[0], components[1]})
			}

			return
		}
	}

	// Empty :(
	return
}

func ask(arns []ARN) (idx int) {
	for {
		for i, a := range arns {
			// Use one-based indexing for human-friendliness.
			fmt.Printf("%d. %s\n", i+1, a.Role)
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
