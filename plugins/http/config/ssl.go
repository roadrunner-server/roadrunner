package config

// SSL defines https server configuration.
type SSL struct {
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
