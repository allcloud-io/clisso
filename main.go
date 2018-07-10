package main

import (
	"log"

	"github.com/allcloud-io/clisso/cmd"
)

// This variable is used by the "version" command and is set during build.
var version = "undefined"

func main() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	cmd.Execute(version)
}
