package http

import (
	"errors"
	"fmt"
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/service"
	"net"
	"os"
	"strings"
)

// Config configures RoadRunner HTTP server.
type Config struct {
	// Port and port to handle as http server.
	Address string

	// SSL defines https server options.
	SSL SSLConfig

	// FCGI configuration. You can use FastCGI without HTTP server.
	FCGI *FCGIConfig

	// HTTP2 configuration
	HTTP2 *HTTP2Config

	// MaxRequestSize specified max size for payload body in megabytes, set 0 to unlimited.
	MaxRequestSize int64

	// TrustedSubnets declare IP subnets which are allowed to set ip using X-Real-Ip and X-Forwarded-For
	TrustedSubnets []string
	cidrs          []*net.IPNet

	// Uploads configures uploads configuration.
	Uploads *UploadsConfig

	// Workers configures rr server and worker pool.
	Workers *roadrunner.ServerConfig
}

// FCGIConfig for FastCGI server.
type FCGIConfig struct {
	// Address and port to handle as http server.
	Address string
}

// HTTP2Config HTTP/2 server customizations.
type HTTP2Config struct {
	// Enable or disable HTTP/2 extension, default enable.
	Enabled bool

	// MaxConcurrentStreams defaults to 128.
	MaxConcurrentStreams uint32
}

// InitDefaults sets default values for HTTP/2 configuration.
func (cfg *HTTP2Config) InitDefaults() error {
	cfg.Enabled = true
	cfg.MaxConcurrentStreams = 128

	return nil
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

// EnableHTTP is true when http server must run.
func (c *Config) EnableHTTP() bool {
	return c.Address != ""
}

// EnableTLS returns true if rr must listen TLS connections.
func (c *Config) EnableTLS() bool {
	return c.SSL.Key != "" || c.SSL.Cert != ""
}

// EnableHTTP2 when HTTP/2 extension must be enabled (only with TSL).
func (c *Config) EnableHTTP2() bool {
	return c.HTTP2.Enabled
}

// EnableFCGI is true when FastCGI server must be enabled.
func (c *Config) EnableFCGI() bool {
	return c.FCGI.Address != ""
}

// Hydrate must populate Config values using given Config source. Must return error if Config is not valid.
func (c *Config) Hydrate(cfg service.Config) error {
	if c.Workers == nil {
		c.Workers = &roadrunner.ServerConfig{}
	}

	if c.HTTP2 == nil {
		c.HTTP2 = &HTTP2Config{}
	}

	if c.FCGI == nil {
		c.FCGI = &FCGIConfig{}
	}

	if c.Uploads == nil {
		c.Uploads = &UploadsConfig{}
	}

	if c.SSL.Port == 0 {
		c.SSL.Port = 443
	}

	c.HTTP2.InitDefaults()
	c.Uploads.InitDefaults()
	c.Workers.InitDefaults()

	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	c.Workers.UpscaleDurations()

	if c.TrustedSubnets == nil {
		// @see https://en.wikipedia.org/wiki/Reserved_IP_addresses
		c.TrustedSubnets = []string{
			"10.0.0.0/8",
			"127.0.0.0/8",
			"172.16.0.0/12",
			"192.168.0.0/16",
			"::1/128",
			"fc00::/7",
			"fe80::/10",
		}
	}

	if err := c.parseCIDRs(); err != nil {
		return err
	}

	return c.Valid()
}

func (c *Config) parseCIDRs() error {
	for _, cidr := range c.TrustedSubnets {
		_, cr, err := net.ParseCIDR(cidr)
		if err != nil {
			return err
		}

		c.cidrs = append(c.cidrs, cr)
	}

	return nil
}

// IsTrusted if api can be trusted to use X-Real-Ip, X-Forwarded-For
func (c *Config) IsTrusted(ip string) bool {
	if c.cidrs == nil {
		return false
	}

	i := net.ParseIP(ip)
	if i == nil {
		return false
	}

	for _, cird := range c.cidrs {
		if cird.Contains(i) {
			return true
		}
	}

	return false
}

// Valid validates the configuration.
func (c *Config) Valid() error {
	if c.Uploads == nil {
		return errors.New("mailformed uploads config")
	}

	if c.HTTP2 == nil {
		return errors.New("mailformed http2 config")
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

	if !c.EnableHTTP() && !c.EnableTLS() && !c.EnableFCGI() {
		return errors.New("unable to run http service, no method has been specified (http, https, http/2 or FastCGI)")
	}

	if c.Address != "" && !strings.Contains(c.Address, ":") {
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
