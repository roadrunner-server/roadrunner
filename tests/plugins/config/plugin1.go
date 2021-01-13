package config

import (
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
)

type AllConfig struct {
	RPC struct {
		Listen string `yaml:"listen"`
	} `yaml:"rpc"`
	Reload struct {
		Enabled  bool     `yaml:"enabled"`
		Interval string   `yaml:"interval"`
		Patterns []string `yaml:"patterns"`
		Services struct {
			HTTP struct {
				Recursive bool     `yaml:"recursive"`
				Ignore    []string `yaml:"ignore"`
				Patterns  []string `yaml:"patterns"`
				Dirs      []string `yaml:"dirs"`
			} `yaml:"http"`
			Jobs struct {
				Recursive bool     `yaml:"recursive"`
				Ignore    []string `yaml:"ignore"`
				Dirs      []string `yaml:"dirs"`
			} `yaml:"jobs"`
			RPC struct {
				Recursive bool     `yaml:"recursive"`
				Patterns  []string `yaml:"patterns"`
				Dirs      []string `yaml:"dirs"`
			} `yaml:"rpc"`
		} `yaml:"services"`
	} `yaml:"reload"`
}

// ReloadConfig is a Reload configuration point.
type ReloadConfig struct {
	Interval time.Duration
	Patterns []string
	Services map[string]ServiceConfig
}

type ServiceConfig struct {
	Enabled   bool
	Recursive bool
	Patterns  []string
	Dirs      []string
	Ignore    []string
}

type Foo struct {
	configProvider config.Configurer
}

// Depends on S2 and DB (S3 in the current case)
func (f *Foo) Init(p config.Configurer) error {
	f.configProvider = p
	return nil
}

func (f *Foo) Serve() chan error {
	const op = errors.Op("foo serve")
	errCh := make(chan error, 1)

	r := &ReloadConfig{}
	err := f.configProvider.UnmarshalKey("reload", r)
	if err != nil {
		errCh <- err
	}

	if len(r.Patterns) == 0 {
		errCh <- errors.E(op, errors.Str("should be at least one pattern, but got 0"))
		return errCh
	}

	var allCfg AllConfig
	err = f.configProvider.Unmarshal(&allCfg)
	if err != nil {
		errCh <- errors.E(op, errors.Str("should be at least one pattern, but got 0"))
		return errCh
	}

	if allCfg.RPC.Listen != "tcp://localhost:6060" {
		errCh <- errors.E(op, errors.Str("RPC.Listen should be parsed"))
		return errCh
	}

	return errCh
}

func (f *Foo) Stop() error {
	return nil
}
