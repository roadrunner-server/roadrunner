package memory

// Config is default config for the in-memory driver
type Config struct {
	// Enabled or disabled (true or false)
	Enabled bool
}

// InitDefaults by default driver is turned off
func (c *Config) InitDefaults() {
	c.Enabled = false
}
