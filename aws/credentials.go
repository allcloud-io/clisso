package aws

import (
	"fmt"
	"io"
)

// Credentials represents a set of temporary credentials received from AWS STS
// (http://docs.aws.amazon.com/STS/latest/APIReference/Welcome.html).
type Credentials struct {
	AccessKeyId     string
	SecretAccessKey string
	SessionToken    string
}

// WriteCredentialsToFile gets a set of temporary AWS credentials and writes them
// to a file.
func WriteCredentialsToFile(c *Credentials, w io.Writer) error {
	b := []byte(fmt.Sprintf("%s\n%s\n%s\n", c.AccessKeyId, c.SecretAccessKey, c.SessionToken))
	_, err := w.Write(b)
	if err != nil {
		return err
	}

	return nil
}

// GetBashCommands gets a set of temporary AWS credentials and returns the Bash
// commands for setting them in the shell.
func GetBashCommands(c *Credentials) string {
	return fmt.Sprintf(
		"export AWS_ACCESS_KEY_ID=%v\nexport AWS_SECRET_ACCESS_KEY=%v\nexport AWS_SESSION_TOKEN=%v\n",
		c.AccessKeyId,
		c.SecretAccessKey,
		c.SessionToken,
	)
}
