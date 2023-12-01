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

	"github.com/go-ini/ini"
	log "github.com/sirupsen/logrus"
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
	log.WithFields(log.Fields{
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
			cfg.DeleteSection(s.Name())
			continue
		}
		log.Tracef("Profile %s expires at %s", s.Name(), v.Format(time.RFC3339))
	}

	return cfg.SaveTo(filename)
}

// WriteToShell writes (prints) credentials to stdout. If windows is true, Windows syntax will be
// used.
func WriteToShell(c *Credentials, windows bool, w io.Writer) {
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

// GetValidCredentials returns profiles which have a aws_expiration key but are not yet expired.
func GetValidCredentials(filename string) ([]Profile, error) {
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
