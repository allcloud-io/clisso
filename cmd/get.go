package cmd

import (
	"github.com/johananl/csso/onelogin"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(cmdGet)
}

var cmdGet = &cobra.Command{
	Use:   "get",
	Short: "Get temporary credentials",
	Long: `Obtain temporary credentials for the currently-selected account by
generating a SAML assertion at the identity provider and using this
assertion to retrieve temporary credentials from the cloud provider.`,
	Run: func(cmd *cobra.Command, args []string) {
		onelogin.Get()
	},
}
