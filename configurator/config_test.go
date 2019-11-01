package configurator_test

import (
	"testing"

	"github.com/NearlyUnique/awsSts/configurator"
	"github.com/stretchr/testify/assert"
)

func Test_an_empty_config_returns_empty_string_values(t *testing.T) {
	var c configurator.Config

	assert.Equal(t, "", c.Value("any-key"))
}

func Test_when_a_source_func_is_values_are_returned_for_a_key(t *testing.T) {
	var c configurator.Config
	var actualKey string

	c.Add(func(k string) string {
		actualKey = k
		return "actual value"
	})

	assert.Equal(t, "actual value", c.Value("any-key"))
	assert.Equal(t, "any-key", actualKey)
}

func Test_when_adding_multiple_sources(t *testing.T) {
	src := func(val string) configurator.Source {
		return func(_ string) string {
			return val
		}
	}
	t.Run("first non-empty value is returned", func(t *testing.T) {
		var c configurator.Config
		c.Add(src(""), src("second src value"))

		assert.Equal(t, "second src value", c.Value("any-key"))
	})
	t.Run("if all return empty strings the result is empty", func(t *testing.T) {
		var c configurator.Config
		c.Add(src(""), src(""))

		assert.Equal(t, "", c.Value("any-key"))
	})
}
