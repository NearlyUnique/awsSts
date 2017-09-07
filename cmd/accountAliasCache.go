package cmd

import (
	"encoding/json"
	"io"
)

type (
	//AccountAliasCache stores previously looked up aliases
	AccountAliasCache struct {
		//Aliases list
		Roles []*RoleAlias `json:"roles"`
	}
	// RoleAlias single account
	RoleAlias struct {
		//Account number
		Role string `json:"role"`
		//Aliases list, typicaly a single entry
		Names []string `json:"names"`
	}
)

func (c *AccountAliasCache) Read(rc io.Reader) error {
	d := json.NewDecoder(rc)
	return d.Decode(c)
}

func (c *AccountAliasCache) Write(wc io.Writer) error {
	e := json.NewEncoder(wc)
	e.SetIndent("", "  ")
	return e.Encode(*c)
}

func (c *AccountAliasCache) findAlias(role string) (*RoleAlias, bool) {
	for i, a := range c.Roles {
		if a.Role == role {
			return c.Roles[i], true
		}
	}
	return nil, false
}

func (c *AccountAliasCache) add(role string, names []*string) {
	var toAdd *RoleAlias
	for i, a := range c.Roles {
		if a.Role == role {
			toAdd = c.Roles[i]
			break
		}
	}
	if toAdd == nil {
		toAdd = &RoleAlias{Role: role}
		c.Roles = append(c.Roles, toAdd)
	}
	for _, n := range names {
		toAdd.Names = append(toAdd.Names, *n)
	}
}

func (a RoleAlias) String() string {
	if len(a.Names) == 0 {
		return ""
	}
	return a.Names[0]
}
