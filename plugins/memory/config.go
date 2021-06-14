package memory

// Config is default config for the in-memory driver
type Config struct {
	// Interval for the check
	Interval int
}

// InitDefaults by default driver is turned off
func (c *Config) InitDefaults() {
	if c.Interval == 0 {
		c.Interval = 60 // seconds
	}
}
