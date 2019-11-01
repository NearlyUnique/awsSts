package configurator

type (
	Source func(string) string
	Config struct {
		sources []Source
	}
)

// Add one or more source func
func (c *Config) Add(src ...Source) {
	for _, s := range src {
		c.sources = append(c.sources, s)
	}
}

// Value for the supplied key from the first non-empty source
func (c *Config) Value(key string) string {
	for _, src := range c.sources {
		if v := src(key); v != "" {
			return v
		}
	}
	return ""
}
