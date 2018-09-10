package env

// Environment aggregates list of environment variables. This interface can be used in custom implementation to drive
// values from external sources.
type Environment interface {
	// GetEnv must return list of env variables.
	GetEnv() (map[string]string, error)

	// SetEnv sets or creates environment value.
	SetEnv(key, value string)
}
