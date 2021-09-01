package memcachedkv

type Config struct {
	// Addr is url for memcached, 11211 port is used by default
	Addr []string
}

func (s *Config) InitDefaults() {
	if s.Addr == nil {
		s.Addr = []string{"127.0.0.1:11211"} // default url for memcached
	}
}
