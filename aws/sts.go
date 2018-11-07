package aws

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

var ErrInvalidSessionDuration = errors.New("InvalidSessionDuration")

// AssumeSAMLRole asumes a Role using the SAMLAssertion specified. If the duration cannot be meet
// it transperently lowers the duration and returns an error in parallel to the valid credentials.
func AssumeSAMLRole(PrincipalArn, RoleArn, SAMLAssertion string, duration int64) (*Credentials, error) {
	creds, err := assumeSAMLRole(PrincipalArn, RoleArn, SAMLAssertion, duration, false)
	if err == ErrInvalidSessionDuration {
		// the requested duration was invalid. So try again with a minimum of 3600s and return an
		// EErrInvalidSessionDuration error, too.
		return assumeSAMLRole(PrincipalArn, RoleArn, SAMLAssertion, 3600, true)
	}

	return creds, nil
}

func assumeSAMLRole(PrincipalArn, RoleArn, SAMLAssertion string, duration int64, roleDoesNotSupportDuration bool) (*Credentials, error) {
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
		// The role might not yet support the requested duration, let's catch and try to lower in 1h
		// steps. There is - as of now - no other way than to do a string comparison.
		if strings.HasPrefix(err.Error(), "ValidationError: The requested DurationSeconds exceeds the MaxSessionDuration set for this role") && duration > 3600 && duration <= 43200 {
			return nil, ErrInvalidSessionDuration
		}
		return nil, fmt.Errorf("assuming role: %v", err)
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

	if roleDoesNotSupportDuration {
		err = ErrInvalidSessionDuration
	}

	return &creds, err
}
