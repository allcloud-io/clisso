package aws

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sts"
)

const stsSuccessXML = `<AssumeRoleWithSAMLResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/"><AssumeRoleWithSAMLResult><Credentials><AccessKeyId>AKIATEST</AccessKeyId><SecretAccessKey>secret</SecretAccessKey><SessionToken>token</SessionToken><Expiration>2030-01-02T03:04:05Z</Expiration></Credentials></AssumeRoleWithSAMLResult><ResponseMetadata><RequestId>req</RequestId></ResponseMetadata></AssumeRoleWithSAMLResponse>`

const stsDurationErrXML = `<ErrorResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/"><Error><Type>Sender</Type><Code>ValidationError</Code><Message>The requested DurationSeconds exceeds the MaxSessionDuration set for this role.</Message></Error><RequestId>req</RequestId></ErrorResponse>`

// fakeSTS installs an STS endpoint override and returns the captured request/region.
func fakeSTS(t *testing.T, status int, body string) (form *url.Values, auth *string, region *string) {
	t.Helper()
	form, auth, region = &url.Values{}, new(string), new(string)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		*form, _ = url.ParseQuery(string(b))
		*auth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	old := STSClientOptions
	STSClientOptions = []func(*sts.Options){func(o *sts.Options) {
		o.BaseEndpoint = &ts.URL
		*region = o.Region
	}}
	t.Cleanup(func() { STSClientOptions = old; ts.Close() })
	return
}

func TestAssumeSAMLRoleSuccess(t *testing.T) {
	form, auth, region := fakeSTS(t, 200, stsSuccessXML)
	creds, err := AssumeSAMLRole("arn:aws:iam::1:saml-provider/P", "arn:aws:iam::1:role/R", "assertion", "eu-west-1", 3600)
	if err != nil {
		t.Fatal(err)
	}
	if creds.AccessKeyID != "AKIATEST" || creds.SecretAccessKey != "secret" || creds.SessionToken != "token" {
		t.Errorf("wrong creds: %+v", creds)
	}
	if !creds.Expiration.Equal(time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)) {
		t.Errorf("wrong expiration: %v", creds.Expiration)
	}
	if form.Get("Action") != "AssumeRoleWithSAML" || form.Get("SAMLAssertion") != "assertion" ||
		form.Get("RoleArn") != "arn:aws:iam::1:role/R" || form.Get("PrincipalArn") != "arn:aws:iam::1:saml-provider/P" ||
		form.Get("DurationSeconds") != "3600" {
		t.Errorf("unexpected request form: %v", *form)
	}
	if *auth != "" {
		t.Errorf("request should be unsigned, got Authorization %q", *auth)
	}
	if *region != "eu-west-1" {
		t.Errorf("region = %q", *region)
	}
}

func TestAssumeSAMLRoleDurationExceeded(t *testing.T) {
	fakeSTS(t, 400, stsDurationErrXML)
	_, err := AssumeSAMLRole("arn:aws:iam::1:saml-provider/P", "arn:aws:iam::1:role/R", "a", "us-east-1", 43200)
	if err == nil || err.Error() != ErrDurationExceeded {
		t.Fatalf("expected %q, got %v", ErrDurationExceeded, err)
	}
}

func TestAssumeSAMLRoleChinaRegion(t *testing.T) {
	_, _, region := fakeSTS(t, 200, stsSuccessXML)
	_, err := AssumeSAMLRole("arn:aws-cn:iam::1:saml-provider/P", "arn:aws-cn:iam::1:role/R", "a", "aws-global", 3600)
	if err != nil {
		t.Fatal(err)
	}
	if *region != "cn-north-1" {
		t.Errorf("region = %q, want cn-north-1", *region)
	}
}
