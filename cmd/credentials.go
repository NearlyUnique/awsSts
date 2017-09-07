package cmd

import (
	"github.com/pkg/errors"
	"github.com/vito/go-interact/interact"
)

// Credentials extract the usrename and password from config or interactivly
func Credentials(username, password string) (string, string, error) {
	var err error
	if len(username) == 0 {
		username, err = usernameFromCli()
	}
	if err == nil && len(password) == 0 {
		password, err = passwordFromCli()
	}
	return username, password, err
}

func usernameFromCli() (string, error) {
	var username = ""
	err := interact.NewInteraction("Username").Resolve(interact.Required(&username))
	if err != nil {
		return "", errors.Wrap(err, "Reading username")
	}
	return string(username), nil
}

func passwordFromCli() (string, error) {
	var password interact.Password
	err := interact.NewInteraction("Password").Resolve(interact.Required(&password))
	if err != nil {
		return "", errors.Wrap(err, "Reading password")
	}
	return string(password), nil
}
