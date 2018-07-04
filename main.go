package main

import (
	"github.com/allcloud-io/clisso/cmd"
)

// This variable is used by the "version" command and is set during build.
var version = "undefined"

func main() {
	cmd.Execute(version)
}
