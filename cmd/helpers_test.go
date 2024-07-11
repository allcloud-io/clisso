package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
