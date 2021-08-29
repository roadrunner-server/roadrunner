package boltjobs

const (
	file     string = "file"
	priority string = "priority"
	prefetch string = "prefetch"
)

type GlobalCfg struct {
	// db file permissions
	Permissions int `mapstructure:"permissions"`
	// consume timeout
}

func (c *GlobalCfg) InitDefaults() {
	if c.Permissions == 0 {
		c.Permissions = 0777
	}
}

type Config struct {
	File     string `mapstructure:"file"`
	Priority int    `mapstructure:"priority"`
	Prefetch int    `mapstructure:"prefetch"`
}

func (c *Config) InitDefaults() {
	if c.File == "" {
		c.File = "rr.db"
	}

	if c.Priority == 0 {
		c.Priority = 10
	}

	if c.Prefetch == 0 {
		c.Prefetch = 1000
	}
}
