package memory

// Config for the memory driver is empty, it's just a placeholder
type Config struct {
	Path string `mapstructure:"path"`
}

func (c *Config) InitDefaults() {}
