package service

import "time"

// Service represents particular service configuration
type Service struct {
	Command          string        `mapstructure:"command"`
	ProcessNum       int           `mapstructure:"process_num"`
	ExecTimeout      time.Duration `mapstructure:"exec_timeout"`
	RestartAfterExit bool          `mapstructure:"restart_after_exit"`
	RestartDelay     time.Duration `mapstructure:"restart_delay"`
}

// Config for the services
type Config struct {
	Services map[string]Service `mapstructure:"service"`
}

func (c *Config) InitDefault() {
	if len(c.Services) > 0 {
		for k, v := range c.Services {
			if v.ProcessNum == 0 {
				val := c.Services[k]
				val.ProcessNum = 1
				c.Services[k] = val
			}
			if v.RestartDelay == 0 {
				val := c.Services[k]
				val.RestartDelay = time.Minute
				c.Services[k] = val
			}
		}
	}
}
