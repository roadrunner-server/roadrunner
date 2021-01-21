package config

// HTTP2 HTTP/2 server customizations.
type HTTP2 struct {
	// h2cHandler is a Handler which implements h2c by hijacking the HTTP/1 traffic
	// that should be h2c traffic. There are two ways to begin a h2c connection
	// (RFC 7540 Section 3.2 and 3.4): (1) Starting with Prior Knowledge - this
	// works by starting an h2c connection with a string of bytes that is valid
	// HTTP/1, but unlikely to occur in practice and (2) Upgrading from HTTP/1 to
	// h2c - this works by using the HTTP/1 Upgrade header to request an upgrade to
	// h2c. When either of those situations occur we hijack the HTTP/1 connection,
	// convert it to a HTTP/2 connection and pass the net.Conn to http2.ServeConn.

	// H2C enables HTTP/2 over TCP
	H2C bool

	// MaxConcurrentStreams defaults to 128.
	MaxConcurrentStreams uint32 `mapstructure:"max_concurrent_streams"`
}

// InitDefaults sets default values for HTTP/2 configuration.
func (cfg *HTTP2) InitDefaults() error {
	if cfg.MaxConcurrentStreams == 0 {
		cfg.MaxConcurrentStreams = 128
	}

	return nil
}
