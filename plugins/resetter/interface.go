package resetter

// Resetter interface
type Resetter interface {
	// Reset reload plugin
	Reset() error
}
