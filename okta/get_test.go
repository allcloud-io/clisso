package okta

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/allcloud-io/clisso/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/go-ini/ini"
	"github.com/spf13/viper"
	keyring "github.com/zalando/go-keyring"
)

const e2eSTSXML = `<AssumeRoleWithSAMLResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/"><AssumeRoleWithSAMLResult><Credentials><AccessKeyId>AKIAE2E</AccessKeyId><SecretAccessKey>e2esecret</SecretAccessKey><SessionToken>e2etoken</SessionToken><Expiration>2030-01-02T03:04:05Z</Expiration></Credentials></AssumeRoleWithSAMLResult><ResponseMetadata><RequestId>r</RequestId></ResponseMetadata></AssumeRoleWithSAMLResponse>`

func runE2E(t *testing.T, mfa bool) {
	assertion, err := os.ReadFile("testdata/assertion")
	if err != nil {
		t.Fatal(err)
	}
	keyring.MockInit()
	if err := keyring.Set("clisso", "e2e-provider", "pw"); err != nil {
		t.Fatal(err)
	}

	var verifyCalls int32
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/authn", func(w http.ResponseWriter, r *http.Request) {
		if mfa {
			fmt.Fprint(w, `{"status":"MFA_REQUIRED","stateToken":"st","_embedded":{"factors":[{"id":"f1","factorType":"push","_links":{"verify":{"href":"x"}}}]}}`)
			return
		}
		fmt.Fprint(w, `{"status":"SUCCESS","sessionToken":"sess"}`)
	})
	mux.HandleFunc("/api/v1/authn/factors/f1/verify", func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&verifyCalls, 1) < 2 {
			fmt.Fprint(w, `{"status":"MFA_CHALLENGE","factorResult":"WAITING"}`)
			return
		}
		fmt.Fprint(w, `{"status":"SUCCESS","sessionToken":"sess"}`)
	})
	mux.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("sessionToken") != "sess" {
			http.Error(w, "no session", 401)
			return
		}
		fmt.Fprintf(w, `<html><body><form id="appForm"><input name="SAMLResponse" value="%s"/></form></body></html>`, strings.TrimSpace(string(assertion)))
	})
	okta := httptest.NewServer(mux)
	defer okta.Close()

	stsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("SAMLAssertion") != strings.TrimSpace(string(assertion)) ||
			r.Form.Get("RoleArn") != "arn:aws:iam::123456789012:role/OneLogin-MyRole" {
			w.WriteHeader(400)
			return
		}
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprint(w, e2eSTSXML)
	}))
	defer stsSrv.Close()
	old := aws.STSClientOptions
	aws.STSClientOptions = []func(*sts.Options){func(o *sts.Options) { o.BaseEndpoint = &stsSrv.URL }}
	defer func() { aws.STSClientOptions = old }()

	viper.Reset()
	defer viper.Reset()
	viper.Set("providers.e2e-provider.base-url", okta.URL)
	viper.Set("providers.e2e-provider.username", "user")
	viper.Set("apps.e2e-app.provider", "e2e-provider")
	viper.Set("apps.e2e-app.url", okta.URL+"/app")

	creds, err := Get("e2e-app", "e2e-provider", "", "us-east-1", 3600, false)
	if err != nil {
		t.Fatal(err)
	}
	if creds.AccessKeyID != "AKIAE2E" || creds.SecretAccessKey != "e2esecret" || creds.SessionToken != "e2etoken" {
		t.Fatalf("unexpected creds %+v", creds)
	}
	if mfa && verifyCalls < 2 {
		t.Errorf("expected polling, verify calls = %d", verifyCalls)
	}

	fn := filepath.Join(t.TempDir(), "credentials")
	if err := aws.OutputFile(creds, fn, "e2e-app"); err != nil {
		t.Fatal(err)
	}
	cfg, err := ini.Load(fn)
	if err != nil {
		t.Fatal(err)
	}
	s := cfg.Section("e2e-app")
	for k, v := range map[string]string{
		"aws_access_key_id":     "AKIAE2E",
		"aws_secret_access_key": "e2esecret",
		"aws_session_token":     "e2etoken",
		"aws_expiration":        time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC).Format(time.RFC3339),
	} {
		if s.Key(k).String() != v {
			t.Errorf("%s = %q, want %q", k, s.Key(k).String(), v)
		}
	}
}

func TestGetEndToEnd(t *testing.T)        { runE2E(t, false) }
func TestGetEndToEndMFAPush(t *testing.T) { runE2E(t, true) }
