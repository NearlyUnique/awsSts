package configurator

import "os"

// EnvSource calls directly to the os environment
var EnvSource Source = os.Getenv

// MapSource returns the value for a key in the supplied map
func MapSource(m map[string]string) Source {
	return func(key string) string {
		return m[key]
	}
}
