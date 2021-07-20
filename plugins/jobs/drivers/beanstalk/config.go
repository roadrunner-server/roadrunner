package beanstalk

import "time"

type GlobalCfg struct {
	Addr    string        `mapstructure:"addr"`
	Timeout time.Duration `mapstructure:"timeout"`
}

func (c *GlobalCfg) InitDefault() {
	if c.Addr == "" {
		c.Addr = "tcp://localhost:11300"
	}

	if c.Timeout == 0 {
		c.Timeout = time.Second * 30
	}
}

type Config struct {
	PipePriority   int64         `mapstructure:"priority"`
	TubePriority   uint32        `mapstructure:"tube_priority"`
	Tube           string        `mapstructure:"tube"`
	ReserveTimeout time.Duration `mapstructure:"reserve_timeout"`
}

func (c *Config) InitDefault() {
	if c.Tube == "" {
		c.Tube = "default"
	}

	if c.ReserveTimeout == 0 {
		c.ReserveTimeout = time.Second * 5
	}

	if c.PipePriority == 0 {
		c.PipePriority = 10
	}
}