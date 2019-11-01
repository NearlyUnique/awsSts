package cmd

import (
	"encoding/base64"

	"github.com/pkg/errors"
	"github.com/vito/go-interact/interact"
)

// Credentials extract the username and password from config or interactively
func Credentials(username, password string) (string, string, error) {
	var err error
	if len(username) == 0 {
		username, err = usernameFromCli()
	}
	if err == nil && len(password) == 0 {
		password, err = passwordFromCli()
	}
	return username, debase64(password), err
}

// debase64 decodes some text if possible, otherwise use as is
func debase64(text string) string {
	dec, err := base64.URLEncoding.DecodeString(text)
	if err == nil {
		return string(dec)
	}
	return text
}

func usernameFromCli() (string, error) {
	var username = ""
	err := interact.NewInteraction("Username").Resolve(interact.Required(&username))
	if err != nil {
		return "", errors.Wrap(err, "Reading username")
	}
	return username, nil
}

func passwordFromCli() (string, error) {
	var password interact.Password
	err := interact.NewInteraction("Password").Resolve(interact.Required(&password))
	if err != nil {
		return "", errors.Wrap(err, "Reading password")
	}
	return string(password), nil
}
