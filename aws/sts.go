package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/fatih/color"
)

// AssumeSAMLRole asumes a Role using the SAMLAssertion specified. If the duration cannot be meet it transperently lowers the duration and sets the returned bool to true to signal that a message should be displayed
func AssumeSAMLRole(PrincipalArn, RoleArn, SAMLAssertion, profile string, duration int64) (*Credentials, error) {
	creds, roleNeedsChange, err := assumeSAMLRole(PrincipalArn, RoleArn, SAMLAssertion, duration, false)
	if err == nil && roleNeedsChange {
		// TODO: This might conflict with spinner... Trying to avoid by starting with \r.
		fmt.Printf(color.YellowString("\rThe role does not support the requested max-session-duration of %v. To have a max session duration for up to 12h run:\n"), duration)
		fmt.Printf("\raws iam update-role --role-name %v --max-session-duration 43200 --profile %v\n", RoleArn[strings.LastIndex(RoleArn, "/")+1:], profile)
	}
	return creds, err
}

func assumeSAMLRole(PrincipalArn, RoleArn, SAMLAssertion string, duration int64, roleDoesNotSupportDuration bool) (*Credentials, bool, error) {
	// Assume role
	input := sts.AssumeRoleWithSAMLInput{
		PrincipalArn:    aws.String(PrincipalArn),
		RoleArn:         aws.String(RoleArn),
		SAMLAssertion:   aws.String(SAMLAssertion),
		DurationSeconds: aws.Int64(duration),
	}

	sess := session.Must(session.NewSession())
	svc := sts.New(sess)

	aResp, err := svc.AssumeRoleWithSAML(&input)
	if err != nil {
		// The role might not yet support the requested duration, let's catch and try to lower in 1h steps
		if strings.HasPrefix(err.Error(), "ValidationError: The requested DurationSeconds exceeds the MaxSessionDuration set for this role") && duration > 3600 && duration <= 43200 {
			duration -= 3600
			return assumeSAMLRole(PrincipalArn, RoleArn, SAMLAssertion, duration, true)
		}
		return nil, false, fmt.Errorf("assuming role: %v", err)
	}

	keyID := *aResp.Credentials.AccessKeyId
	secretKey := *aResp.Credentials.SecretAccessKey
	sessionToken := *aResp.Credentials.SessionToken
	expiration := *aResp.Credentials.Expiration

	creds := Credentials{
		AccessKeyID:     keyID,
		SecretAccessKey: secretKey,
		SessionToken:    sessionToken,
		Expiration:      expiration,
	}

	return &creds, roleDoesNotSupportDuration, nil
}
