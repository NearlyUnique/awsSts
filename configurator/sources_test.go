package configurator_test

import (
	"os"
	"testing"

	"github.com/NearlyUnique/awsSts/configurator"
	"github.com/stretchr/testify/assert"
)

func Test_env_source_read_from_os_environment(t *testing.T) {
	assert.NoError(t, os.Setenv("ant_environment_value_for_config_test", "the value"))
	assert.NoError(t, os.Unsetenv("any_unknown_env_var"))

	assert.Equal(t, "the value", configurator.EnvSource("ant_environment_value_for_config_test"))
	assert.Equal(t, "", configurator.EnvSource("any_unknown_env_var"))
}

func Test_map_source_read_from_a_string_map(t *testing.T) {
	m := map[string]string{
		"the-key":    "the map value",
		"unused-key": "unused value",
	}

	mapSrc := configurator.MapSource(m)

	assert.Equal(t, "the map value", mapSrc("the-key"))
	assert.Equal(t, "", mapSrc("unknown-the-key"))
}
