package resetter

// If plugin implements Resettable interface, than it state can be resetted without reload in runtime via RPC/HTTP
type Resettable interface {
	// Reset reload all plugins
	Reset() error
}

// Resetter interface is the Resetter plugin main interface
type Resetter interface {
	// Reset all registered plugins
	ResetAll() error
	// Reset by plugin name
	ResetByName(string) error
	// GetAll registered plugins
	GetAll() []string
}
