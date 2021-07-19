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

type Config struct{}

func (c *Config) InitDefault() {}
