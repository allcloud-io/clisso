package saml

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
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
		err = fmt.Errorf("No valid AWS roles were returned")

		return

	case 1:
		a = arns[0]

		return
	}

	// Many ARNs returned, ask
	for {
		idx, err := ask(arns)
		if err != nil {
			fmt.Println(err.Error())
		} else if idx < 0 || idx > len(arns) {
			fmt.Printf("%d is an invalid value. Valid values are 0 to %d inclusive", idx, len(arns)-1)
		} else {
			a = arns[idx]

			break
		}
	}

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
				components := strings.Split(av.Value, ",")

				arns = append(arns, ARN{components[0], components[1]})
			}

			return
		}
	}

	// Empty :(
	return
}

func ask(arns []ARN) (idx int, err error) {
	for idx, a := range arns {
		fmt.Printf("%d. %s\n", idx, a.Role)
	}

	fmt.Print("Please select an ARN to assume: ")

	_, err = fmt.Scanln(&idx)
	return
}
