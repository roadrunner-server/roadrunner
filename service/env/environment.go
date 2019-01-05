package env

// Environment aggregates list of environment variables. This interface can be used in custom implementation to drive
// values from external sources.
type Environment interface {
	Setter
	Getter

	// Copy all environment values.
	Copy(setter Setter) error
}

// Setter provides ability to set environment value.
type Setter interface {
	// SetEnv sets or creates environment value.
	SetEnv(key, value string)
}

// Getter provides ability to set environment value.
type Getter interface {
	// GetEnv must return list of env variables.
	GetEnv() (map[string]string, error)
}
