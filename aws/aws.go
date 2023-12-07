/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package aws

import (
	"fmt"
	"io"
	"time"

	"github.com/allcloud-io/clisso/log"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
)

// Credentials represents a set of temporary credentials received from AWS STS
// (http://docs.aws.amazon.com/STS/latest/APIReference/Welcome.html).
type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Expiration      time.Time
}

// Profile represents an AWS profile
type Profile struct {
	Name         string
	LifetimeLeft time.Duration
	ExpireAtUnix int64
}

const expireKey = "aws_expiration"

// WriteToFile writes credentials to an AWS CLI credentials file
// (https://docs.aws.amazon.com/cli/latest/userguide/cli-config-files.html). In addition, this
// function removes expired temporary credentials from the credentials file.
func WriteToFile(c *Credentials, filename string, section string) error {
	log.Log.WithFields(logrus.Fields{
		"filename": filename,
		"section":  section,
	}).Debug("Writing credentials to file")
	cfg, err := ini.LooseLoad(filename)
	if err != nil {
		return err
	}
	cfg.DeleteSection(section)
	_, err = cfg.Section(section).NewKey("aws_access_key_id", c.AccessKeyID)
	if err != nil {
		return err
	}
	_, err = cfg.Section(section).NewKey("aws_secret_access_key", c.SecretAccessKey)
	if err != nil {
		return err
	}
	_, err = cfg.Section(section).NewKey("aws_session_token", c.SessionToken)
	if err != nil {
		return err
	}
	_, err = cfg.Section(section).NewKey(expireKey, c.Expiration.UTC().Format(time.RFC3339))
	if err != nil {
		return err
	}

	// Remove expired credentials.
	for _, s := range cfg.Sections() {
		if !s.HasKey(expireKey) {
			log.Log.Tracef("Skipping profile %s because it does not have an %s key", s.Name(), expireKey)
			continue
		}
		v, err := s.Key(expireKey).TimeFormat(time.RFC3339)
		if err != nil {
			log.Log.Warnf("Cannot parse date (%v) in profile %s: %s",
				s.Key(expireKey), s.Name(), err)
			continue
		}
		if time.Now().UTC().Unix() > v.Unix() {
			log.Log.Tracef("Removing expired credentials for profile %s", s.Name())
			cfg.DeleteSection(s.Name())
			continue
		}
		log.Log.Tracef("Profile %s expires at %s", s.Name(), v.Format(time.RFC3339))
	}

	return cfg.SaveTo(filename)
}

// WriteToStdOutAsEnvironment writes (prints) credentials to stdout. If windows is true, Windows syntax will be
// used. The output can be used to set environment variables.
func WriteToStdOutAsEnvironment(c *Credentials, windows bool, w io.Writer) {
	fmt.Print("Please paste the following in your shell:")
	if windows {
		fmt.Fprintf(
			w,
			"set AWS_ACCESS_KEY_ID=%v\nset AWS_SECRET_ACCESS_KEY=%v\nset AWS_SESSION_TOKEN=%v\n",
			c.AccessKeyID,
			c.SecretAccessKey,
			c.SessionToken,
		)
	} else {
		fmt.Fprintf(
			w,
			"export AWS_ACCESS_KEY_ID=%v\nexport AWS_SECRET_ACCESS_KEY=%v\nexport AWS_SESSION_TOKEN=%v\n",
			c.AccessKeyID,
			c.SecretAccessKey,
			c.SessionToken,
		)
	}
}

// WriteCredentialsToStdOutAsCredentialProcess writes (prints) credentials to stdout in the format required by the AWS CLI.
// The output can be used to set the credential_process option in the AWS CLI configuration file.
func WriteCredentialsToStdOutAsCredentialProcess(c *Credentials, w io.Writer) {
	log.Log.Trace("Writing credentials to stdout in credential_process format")
	log.Log.Infof("Credentials expire at %s, in %d Minutes", c.Expiration.Format(time.RFC3339), int(c.Expiration.Sub(time.Now().UTC()).Minutes()))
	fmt.Fprintf(
		w,
		`{ "Version": 1, "AccessKeyId": %q, "SecretAccessKey": %q, "SessionToken": %q, "Expiration": %q }`,
		c.AccessKeyID,
		c.SecretAccessKey,
		c.SessionToken,
		// Time must be in ISO8601 format
		c.Expiration.Format(time.RFC3339),
	)
}

// GetValidProfiles returns profiles which have a aws_expiration key but are not yet expired.
func GetValidProfiles(filename string) ([]Profile, error) {
	var profiles []Profile
	log.Log.WithField("filename", filename).Trace("Loading AWS credentials file")
	cfg, err := ini.LooseLoad(filename)
	if err != nil {
		err = fmt.Errorf("%s contains errors: %w", filename, err)
		log.Log.WithError(err).Trace("Failed to load AWS credentials file")
		return nil, err
	}
	for _, s := range cfg.Sections() {
		if s.HasKey(expireKey) {
			v, err := s.Key(expireKey).TimeFormat(time.RFC3339)
			if err != nil {
				log.Log.Warnf("Cannot parse date (%v) in section %s: %s",
					s.Key(expireKey), s.Name(), err)
				continue
			}

			if time.Now().UTC().Unix() < v.Unix() {
				profile := Profile{Name: s.Name(), ExpireAtUnix: v.Unix(), LifetimeLeft: v.Sub(time.Now().UTC())}
				profiles = append(profiles, profile)
			}

		}
	}
	return profiles, nil
}

// GetValidCredentials returns credentials which have a aws_expiration key but are not yet expired.
// returns a map of profile name to credentials
func GetValidCredentials(filename string) (map[string]Credentials, error) {
	credentials := make(map[string]Credentials)
	log.Log.WithField("filename", filename).Trace("Loading credentials file")
	cfg, err := ini.LooseLoad(filename)
	if err != nil {
		err = fmt.Errorf("%s contains errors: %w", filename, err)
		log.Log.WithError(err).Trace("Failed to load credentials file")
		return nil, err
	}
	for _, s := range cfg.Sections() {
		if s.HasKey(expireKey) {
			v, err := s.Key(expireKey).TimeFormat(time.RFC3339)
			if err != nil {
				log.Log.Warnf("Cannot parse date (%v) in section %s: %s",
					s.Key(expireKey), s.Name(), err)
				continue
			}

			if time.Now().UTC().Unix() < v.Unix() {
				credential := Credentials{
					AccessKeyID:     s.Key("aws_access_key_id").String(),
					SecretAccessKey: s.Key("aws_secret_access_key").String(),
					SessionToken:    s.Key("aws_session_token").String(),
					Expiration:      v,
				}
				credentials[s.Name()] = credential
			}

		}
	}
	return credentials, nil
}
