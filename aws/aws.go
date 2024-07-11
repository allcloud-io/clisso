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

func validateSection(cfg *ini.File, section string) error {
	// if it doesn't exist, we're good
	if cfg.Section(section) == nil {
		return nil
	}
	s := cfg.Section(section)
	// it should not have any of source_profile, role_arn, mfa_serial, external_id, or credential_source
	for _, key := range []string{"source_profile", "role_arn", "mfa_serial", "external_id", "credential_source", "credential_process"} {
		if s.HasKey(key) {
			log.WithFields(log.Fields{
				"section": section,
				"key":     key,
			}).Errorf("Profile contains key %s, which indicates, it should not be used by clisso", key)
			return fmt.Errorf("profile %s contains key %s, which indicates, it should not be used by clisso", section, key)
		}
	}
	return nil
}

// OutputFile writes credentials to an AWS CLI credentials file
// (https://docs.aws.amazon.com/cli/latest/userguide/cli-config-files.html). In addition, this
// function removes expired temporary credentials from the credentials file.
func OutputFile(c *Credentials, filename string, section string) error {
	log.WithFields(log.Fields{
		"filename": filename,
		"section":  section,
	}).Debug("Writing credentials to file")
	cfg, err := ini.LooseLoad(filename)
	if err != nil {
		return err
	}
	err = validateSection(cfg, section)
	if err != nil {
		return err
	}
	if cfg.HasSection(section) {
		log.Tracef("Section %s exists and has passed validation, adding aws_access_key_id, aws_secret_access_key, aws_session_token, %s keys to it", section, expireKey)
	}

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
			log.Tracef("Skipping profile %s because it does not have an %s key", s.Name(), expireKey)
			continue
		}
		v, err := s.Key(expireKey).TimeFormat(time.RFC3339)
		if err != nil {
			log.Warnf("Cannot parse date (%v) in profile %s: %s",
				s.Key(expireKey), s.Name(), err)
			continue
		}
		if time.Now().UTC().Unix() > v.Unix() {
			log.Tracef("Removing expired credentials for profile %s", s.Name())
			for _, key := range []string{"aws_access_key_id", "aws_secret_access_key", "aws_session_token", expireKey} {
				cfg.Section(s.Name()).DeleteKey(key)
			}
			if len(cfg.Section(s.Name()).Keys()) == 0 {
				log.Tracef("Removing empty profile %s", s.Name())
				cfg.DeleteSection(s.Name())
			}
			continue
		}
		log.Tracef("Profile %s expires at %s", s.Name(), v.Format(time.RFC3339))
	}

	return cfg.SaveTo(filename)
}

// OutputEnvironment writes (prints) credentials to stdout. If windows is true, Windows syntax will be
// used. The output can be used to set environment variables.
func OutputEnvironment(c *Credentials, windows bool, w io.Writer) {
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

// OutputCredentialProcess writes (prints) credentials to stdout in the format required by the AWS CLI.
// The output can be used to set the credential_process option in the AWS CLI configuration file.
func OutputCredentialProcess(c *Credentials, w io.Writer) {
	log.Trace("Writing credentials to stdout in credential_process format")
	log.Infof("Credentials expire at %s, in %d Minutes", c.Expiration.Format(time.RFC3339), int(c.Expiration.Sub(time.Now().UTC()).Minutes()))
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
	log.WithField("filename", filename).Trace("Loading AWS credentials file")
	cfg, err := ini.LooseLoad(filename)
	if err != nil {
		err = fmt.Errorf("%s contains errors: %w", filename, err)
		log.WithError(err).Trace("Failed to load AWS credentials file")
		return nil, err
	}
	for _, s := range cfg.Sections() {
		if s.HasKey(expireKey) {
			v, err := s.Key(expireKey).TimeFormat(time.RFC3339)
			if err != nil {
				log.Warnf("Cannot parse date (%v) in section %s: %s",
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
	log.WithField("filename", filename).Trace("Loading credentials file")
	cfg, err := ini.LooseLoad(filename)
	if err != nil {
		err = fmt.Errorf("%s contains errors: %w", filename, err)
		log.WithError(err).Trace("Failed to load credentials file")
		return nil, err
	}
	for _, s := range cfg.Sections() {
		if s.HasKey(expireKey) {
			v, err := s.Key(expireKey).TimeFormat(time.RFC3339)
			if err != nil {
				log.Warnf("Cannot parse date (%v) in section %s: %s",
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
