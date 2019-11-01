package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/vito/go-interact/interact"
)

// SelectRole to create tokens for
func SelectRole(defaultRole string, roles []Arn) (*Arn, error) {
	if len(roles) == 1 {
		journal("Using Role %q\n", roles[0].role)
		return &roles[0], nil
	}
	var list []interact.Choice
	list = append(list, interact.Choice{Display: "Quit", Value: 0})

	for i, a := range roles {
		if a.role == defaultRole || a.alias == defaultRole {
			journal("Default role found %q\n", a.roleMenu())
			return &a, nil
		}
		list = append(list, interact.Choice{Display: a.roleMenu(), Value: i + 1})
	}
	if len(defaultRole) > 0 {
		journal("Default role not found %q", defaultRole)
	}

	choice := 1
	err := interact.NewInteraction(
		"Choose a role",
		list...,
	).Resolve(&choice)
	if choice == 0 {
		journal("Canceled by user")
		os.Exit(0)
	}
	if err != nil {
		return nil, errors.Wrap(err, "No choice")
	}
	return &roles[choice-1], nil
}
