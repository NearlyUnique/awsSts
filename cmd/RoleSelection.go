package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/vito/go-interact/interact"
)

// SelectRole to create tokens for
func SelectRole(defaultRole string, roles []Arn) (*Arn, error) {
	list, autoRole := autoSelectRole(defaultRole, roles)
	if autoRole != nil {
		return autoRole, nil
	}
	if len(list) == 0 {
		return nil, errors.New("No roles available")
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

func autoSelectRole(defaultRole string, roles []Arn) ([]interact.Choice, *Arn) {
	if len(roles) == 0 {
		journal("No Roles available\n")
		return nil, nil
	}

	if len(roles) == 1 {
		journal("Using Role %q\n", roles[0].role)
		return nil, &roles[0]
	}
	var list []interact.Choice
	list = append(list, interact.Choice{Display: "Quit", Value: 0})

	var autoRole Arn
	autoRoleCount := 0

	for i, a := range roles {
		if a.role == defaultRole || a.alias == defaultRole {
			autoRole = a
			autoRoleCount++
		}
		list = append(list, interact.Choice{Display: a.roleMenu(), Value: i + 1})
	}
	if autoRoleCount == 1 {
		journal("Default role found %q\n", autoRole.roleMenu())
		return nil, &autoRole
	}
	if len(defaultRole) > 0 {
		journal("Default role not found %q", defaultRole)
	}
	return list, nil
}
