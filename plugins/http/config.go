package http

import (
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spiral/errors"
	poolImpl "github.com/spiral/roadrunner/v2/pkg/pool"
)

type Cidrs []*net.IPNet

func (c *Cidrs) IsTrusted(ip string) bool {
	if len(*c) == 0 {
		return false
	}

	i := net.ParseIP(ip)
	if i == nil {
		return false
	}

	for _, cird := range *c {
		if cird.Contains(i) {
			return true
		}
	}

	return false
}

// Config configures RoadRunner HTTP server.
type Config struct {
	// Port and port to handle as http server.
	Address string

	// SSL defines https server options.
	SSL *SSLConfig

	// FCGI configuration. You can use FastCGI without HTTP server.
	FCGI *FCGIConfig

	// HTTP2 configuration
	HTTP2 *HTTP2Config

	// MaxRequestSize specified max size for payload body in megabytes, set 0 to unlimited.
	MaxRequestSize uint64

	// TrustedSubnets declare IP subnets which are allowed to set ip using X-Real-Ip and X-Forwarded-For
	TrustedSubnets []string

	// Uploads configures uploads configuration.
	Uploads *UploadsConfig

	// Pool configures worker pool.
	Pool *poolImpl.Config

	// Env is environment variables passed to the  http pool
	Env map[string]string

	// List of the middleware names (order will be preserved)
	Middleware []string

	// slice of net.IPNet
	cidrs Cidrs
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

	// H2C enables HTTP/2 over TCP
	H2C bool

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

	// Root CA file
	RootCA string
}

// EnableHTTP is true when http server must run.
func (c *Config) EnableHTTP() bool {
	return c.Address != ""
}

// EnableTLS returns true if pool must listen TLS connections.
func (c *Config) EnableTLS() bool {
	return c.SSL.Key != "" || c.SSL.Cert != "" || c.SSL.RootCA != ""
}

// EnableHTTP2 when HTTP/2 extension must be enabled (only with TSL).
func (c *Config) EnableHTTP2() bool {
	return c.HTTP2.Enabled
}

// EnableH2C when HTTP/2 extension must be enabled on TCP.
func (c *Config) EnableH2C() bool {
	return c.HTTP2.H2C
}

// EnableFCGI is true when FastCGI server must be enabled.
func (c *Config) EnableFCGI() bool {
	return c.FCGI.Address != ""
}

// Hydrate must populate Config values using given Config source. Must return error if Config is not valid.
func (c *Config) InitDefaults() error {
	if c.Pool == nil {
		// default pool
		c.Pool = &poolImpl.Config{
			Debug:           false,
			NumWorkers:      int64(runtime.NumCPU()),
			MaxJobs:         1000,
			AllocateTimeout: time.Second * 60,
			DestroyTimeout:  time.Second * 60,
			Supervisor:      nil,
		}
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

	if c.SSL == nil {
		c.SSL = &SSLConfig{}
	}

	if c.SSL.Port == 0 {
		c.SSL.Port = 443
	}

	err := c.HTTP2.InitDefaults()
	if err != nil {
		return err
	}
	err = c.Uploads.InitDefaults()
	if err != nil {
		return err
	}

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

	cidrs, err := ParseCIDRs(c.TrustedSubnets)
	if err != nil {
		return err
	}
	c.cidrs = cidrs

	return c.Valid()
}

func ParseCIDRs(subnets []string) (Cidrs, error) {
	c := make(Cidrs, 0, len(subnets))
	for _, cidr := range subnets {
		_, cr, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}

		c = append(c, cr)
	}

	return c, nil
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
	const op = errors.Op("validation")
	if c.Uploads == nil {
		return errors.E(op, errors.Str("malformed uploads config"))
	}

	if c.HTTP2 == nil {
		return errors.E(op, errors.Str("malformed http2 config"))
	}

	if c.Pool == nil {
		return errors.E(op, "malformed pool config")
	}

	if !c.EnableHTTP() && !c.EnableTLS() && !c.EnableFCGI() {
		return errors.E(op, errors.Str("unable to run http service, no method has been specified (http, https, http/2 or FastCGI)"))
	}

	if c.Address != "" && !strings.Contains(c.Address, ":") {
		return errors.E(op, errors.Str("malformed http server address"))
	}

	if c.EnableTLS() {
		if _, err := os.Stat(c.SSL.Key); err != nil {
			if os.IsNotExist(err) {
				return errors.E(op, errors.Errorf("key file '%s' does not exists", c.SSL.Key))
			}

			return err
		}

		if _, err := os.Stat(c.SSL.Cert); err != nil {
			if os.IsNotExist(err) {
				return errors.E(op, errors.Errorf("cert file '%s' does not exists", c.SSL.Cert))
			}

			return err
		}

		// RootCA is optional, but if provided - check it
		if c.SSL.RootCA != "" {
			if _, err := os.Stat(c.SSL.RootCA); err != nil {
				if os.IsNotExist(err) {
					return errors.E(op, errors.Errorf("root ca path provided, but path '%s' does not exists", c.SSL.RootCA))
				}
				return err
			}
		}
	}

	return nil
}
