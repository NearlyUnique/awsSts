package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_if_no_roles_then_empty_list_and_none_selected(t *testing.T) {
	list, role := autoSelectRole("", nil)

	assert.Nil(t, role)
	assert.Empty(t, list)
}

func Test_if_only_one_role_then_it_is_selected(t *testing.T) {
	list, role := autoSelectRole("", []Arn{
		{role: "an_arn"},
	})

	require.NotNil(t, role)
	assert.Equal(t, "an_arn", role.role)
	assert.Empty(t, list)
}

func Test_if_only_one_role_but_does_not_match_default_then_none_are_selected(t *testing.T) {
	list, role := autoSelectRole("unknown-role", []Arn{
		{role: "an_arn"},
	})

	require.NotNil(t, role)
	assert.Equal(t, "an_arn", role.role)
	assert.Empty(t, list)
}

func Test_if_aliases_are_non_unique_matching_default_none_selected(t *testing.T) {
	list, role := autoSelectRole("common-alias", []Arn{
		{role: "an_arn1", alias: "common-alias"},
		{role: "an_arn2", alias: "unique-alias"},
		{role: "an_arn3", alias: "common-alias"},
	})

	assert.Nil(t, role)
	assert.Equal(t, 4, len(list))
}

func Test_if_aliases_are_non_unique_selected_can_be_made_by_arn(t *testing.T) {
	list, role := autoSelectRole("an_arn3", []Arn{
		{role: "an_arn1", alias: "common-alias"},
		{role: "an_arn2", alias: "unique-alias"},
		{role: "an_arn3", alias: "common-alias"},
	})

	require.NotNil(t, role)
	assert.Equal(t, "an_arn3", role.role)
	assert.Empty(t, list)
}

func Test_if_aliases_are_unique_roles_can_be_selected_by_alias(t *testing.T) {
	list, role := autoSelectRole("alias2", []Arn{
		{role: "an_arn1", alias: "alias1"},
		{role: "an_arn2", alias: "alias2"},
		{role: "an_arn3", alias: "alias3"},
	})

	require.NotNil(t, role)
	assert.Equal(t, "an_arn2", role.role)
	assert.Equal(t, "alias2", role.alias)
	assert.Empty(t, list)
}
