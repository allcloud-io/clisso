// +build windows

package keychain

import (
	"errors"
	"log"

	"github.com/fatih/color"
)

func set(provider string, password []byte) (err error) {
	log.Fatal(color.RedString("Storing passwords is not supported on windows"))
	return
}

func get(provider string) (pw []byte, err error) {
	return nil, errors.New("Platform is not supported yet")
}
