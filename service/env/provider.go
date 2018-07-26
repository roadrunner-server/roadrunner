package env

// Provider aggregates list of environment variables. This interface can be used in custom implementation to drive
// values from external sources.
type Provider interface {
	// GetEnv must return list of env variables.
	GetEnv() map[string]string
}
