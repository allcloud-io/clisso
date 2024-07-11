/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */
package cmd

import (
	"testing"

	"github.com/allcloud-io/clisso/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var _, _ = log.SetupLogger("panic", "", false, true)

var testdata = []struct {
	app      int32
	provider int32
	result   int32
}{
	{0, 0, 3600},
	{7200, 0, 7200},
	{0, 7200, 7200},
	{7200, 14400, 7200},
}

func TestSessionDuration(t *testing.T) {
	for _, tc := range testdata {
		viper.Set("apps.test.duration", tc.app)
		viper.Set("providers.test.duration", tc.provider)

		res := sessionDuration("test", "test")
		if res != tc.result {
			t.Fatalf("Invalid duration: got %v, want: %v", res, tc.result)
		}
	}
}

func TestPreferredOutput(t *testing.T) {
	testdata := []struct {
		outputFlag      string
		writeToFileFlag string
		appConfig       string
		globalConfig    string
		result          string
	}{
		{"environment", "", "", "", "environment"},
		{"credential_process", "", "", "", "credential_process"},
		{"", "", "~/.aws/test", "", "~/.aws/test"},
		{"", "", "", "test", "test"},
		{defaultOutput, "", "", "", defaultOutput},
		{defaultOutput, "", "credential_process", "", "credential_process"},
		{defaultOutput, "", "credential_process", "~/global", "credential_process"},
		{defaultOutput, "", "", "~/global", "~/global"},
		{"~/test", "", "credential_process", "", "~/test"},
		{"~/test", "", "credential_process", "~/global", "~/test"},
		{"~/test", "", "", "~/global", "~/test"},
	}
	for _, tc := range testdata {
		viper.Set("apps.test.output", tc.appConfig)
		viper.Set("global.output", tc.globalConfig)

		cmd := &cobra.Command{}
		cmd.Flags().StringVarP(
			&output, "output", "o", tc.outputFlag, "fake",
		)
		cmd.Flags().StringVarP(
			&output, "write-to-file", "f", tc.outputFlag, "fake legacy flag",
		)

		res := preferredOutput(cmd, "test")
		if res != tc.result {
			t.Fatalf("Invalid output: got %v, want: %v", res, tc.result)
		}
	}
}
