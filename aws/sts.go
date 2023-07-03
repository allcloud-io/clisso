/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package aws

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
)

const (
	// A friendly message to show to the user when a requested duration exceeds the configured
	// maximum.
	DurationExceededMessage = "The requested duration exceeded the allowed maximum. Falling " +
		"back to 1 hour.\nTo update the maximum session duration you can use the following " +
		"command:\n\naws iam update-role --role-name <role_name> --max-session-duration " +
		"<duration>\n\nFor more information please refer to the AWS documentation:\n" +
		"https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_manage_modify.html"
	// The error message STS returns when attempting to assume a role with a duration longer than
	// the configured maximum for that role.
	ErrInvalidSessionDuration = "The requested DurationSeconds exceeds the MaxSessionDuration " +
		"set for this role."
	// A custom error which indicates that the requested duration exceeded the configured maximum.
	// TODO Replace this with a custom error type.
	ErrDurationExceeded = "DurationExceeded"
)

// AssumeSAMLRole assumes an AWS IAM role using a SAML assertion.
// In cases where the requested session duration is higher than the maximum allowed on AWS, STS
// returns a specific error message to indicate that. In this case we return a custom error to the
// caller to allow special handling such as retrying with a lower duration.
func AssumeSAMLRole(PrincipalArn, RoleArn, SAMLAssertion, awsRegion string, duration int32) (*Credentials, error) {
	creds, err := assumeSAMLRole(PrincipalArn, RoleArn, SAMLAssertion, awsRegion, duration)
	if err != nil {
		// Check if API error returned by AWS
		var ae smithy.APIError
		if errors.As(err, &ae) {
			// Check if error indicates exceeded duration, no structured error exists so check error message content.
			if strings.Contains(ae.ErrorMessage(), "'durationSeconds' failed to satisfy constraint") || ae.ErrorMessage() == ErrInvalidSessionDuration {
				// Return a custom error to allow the caller to retry etc.
				// TODO Return a custom error type instead of a special value:
				// https://dave.cheney.net/2014/12/24/inspecting-errors
				return nil, errors.New(ErrDurationExceeded)
			}

		}
		return nil, err
	}

	return creds, nil
}

func assumeSAMLRole(PrincipalArn, RoleArn, SAMLAssertion, awsRegion string, duration int32) (*Credentials, error) {
	input := sts.AssumeRoleWithSAMLInput{
		PrincipalArn:    aws.String(PrincipalArn),
		RoleArn:         aws.String(RoleArn),
		SAMLAssertion:   aws.String(SAMLAssertion),
		DurationSeconds: aws.Int32(duration),
	}

	ctx := context.Background()

	config, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		return nil, err
	}

	// If we request credentials for China we need to provide a Chinese region
	idp := regexp.MustCompile(`^arn:aws-cn:iam::\d+:saml-provider\/\S+$`)
	if idp.MatchString(PrincipalArn) && !strings.HasPrefix(awsRegion, "cn-") {
		config.Region = "cn-north-1"
	}
	svc := sts.NewFromConfig(config)

	aResp, err := svc.AssumeRoleWithSAML(ctx, &input)
	if err != nil {
		return nil, err
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

	return &creds, nil
}
