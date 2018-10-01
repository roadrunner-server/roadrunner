package http

import (
	"errors"
	"fmt"
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/service"
	"os"
	"strings"
)

// Config configures RoadRunner HTTP server.
type Config struct {
	// Port and port to handle as http server.
	Address string

	// SSL defines https server options.
	SSL SSLConfig

	// MaxRequest specified max size for payload body in megabytes, set 0 to unlimited.
	MaxRequest int64

	// Uploads configures uploads configuration.
	Uploads *UploadsConfig

	// Workers configures roadrunner server and worker pool.
	Workers *roadrunner.ServerConfig
}

// SSLConfig defines https server configuration.
type SSLConfig struct {
	// Port to listen as HTTPS server, defaults to 443.
	Port int

	// Redirect when enabled forces all http connections to switch to https.
	Redirect bool

	// Key defined private server key.
	Key string

	// Cert is https certificate.
	Cert string
}

// EnableTLS returns true if rr must listen TLS connections.
func (c *Config) EnableTLS() bool {
	return c.SSL.Key != "" || c.SSL.Cert != ""
}

// Hydrate must populate Config values using given Config source. Must return error if Config is not valid.
func (c *Config) Hydrate(cfg service.Config) error {
	if c.Workers == nil {
		c.Workers = &roadrunner.ServerConfig{}
	}

	if c.Uploads == nil {
		c.Uploads = &UploadsConfig{}
	}

	if c.SSL.Port == 0 {
		c.SSL.Port = 443
	}

	c.Uploads.InitDefaults()
	c.Workers.InitDefaults()

	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	c.Workers.UpscaleDurations()

	return c.Valid()
}

// Valid validates the configuration.
func (c *Config) Valid() error {
	if c.Uploads == nil {
		return errors.New("mailformed uploads config")
	}

	if c.Workers == nil {
		return errors.New("mailformed workers config")
	}

	if c.Workers.Pool == nil {
		return errors.New("mailformed workers config (pool config is missing)")
	}

	if err := c.Workers.Pool.Valid(); err != nil {
		return err
	}

	if !strings.Contains(c.Address, ":") {
		return errors.New("mailformed http server address")
	}

	if c.EnableTLS() {
		if _, err := os.Stat(c.SSL.Key); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("key file '%s' does not exists", c.SSL.Key)
			}

			return err
		}

		if _, err := os.Stat(c.SSL.Cert); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("cert file '%s' does not exists", c.SSL.Cert)
			}

			return err
		}
	}

	return nil
}
