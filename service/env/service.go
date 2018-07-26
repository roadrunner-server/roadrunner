package env

// ID contains default svc name.
const ID = "env"

// Service provides ability to map _ENV values from config file.
type Service struct {
	cfg *Config
}

// Init must return configure svc and return true if svc hasStatus enabled. Must return error in case of
// misconfiguration. Services must not be used without proper configuration pushed first.
func (s *Service) Init(cfg *Config) (bool, error) {
	s.cfg = cfg
	return true, nil
}

// GetEnv must return list of env variables.
func (s *Service) GetEnv() map[string]string {
	return s.cfg.Values
}
