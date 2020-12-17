package redis

type Config struct {
	// Addr is address to use. If len > 1, cluster client will be used
	Addr []string
	// database number to use, 0 is used by default
	DB int
	// Master name for failover client, empty by default
	Master string
	// Redis password, empty by default
	Password string
}

// InitDefaults initializing fill config with default values
func (s *Config) InitDefaults() error {
	s.Addr = []string{"localhost:6379"} // default addr is pointing to local storage
	return nil
}
