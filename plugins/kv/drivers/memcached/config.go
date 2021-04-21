package memcached

type Config struct {
	// Addr is url for memcached, 11211 port is used by default
	Addr []string
}

func (s *Config) InitDefaults() {
	if s.Addr == nil {
		s.Addr = []string{"localhost:11211"} // default url for memcached
	}
}
