package env

// ID contains default service name.
const ID = "env"

// Service provides ability to map _ENV values from config file.
type Service struct {
	// values is default set of values.
	values map[string]string
}

// NewService creates new env service instance for given rr version.
func NewService(defaults map[string]string) *Service {
	s := &Service{values: defaults}
	return s
}

// Init must return configure svc and return true if svc hasStatus enabled. Must return error in case of
// misconfiguration. Services must not be used without proper configuration pushed first.
func (s *Service) Init(cfg *Config) (bool, error) {
	if s.values == nil {
		s.values = make(map[string]string)
		s.values["RR"] = "true"
	}

	for k, v := range cfg.Values {
		s.values[k] = v
	}

	return true, nil
}

// GetEnv must return list of env variables.
func (s *Service) GetEnv() (map[string]string, error) {
	return s.values, nil
}

// SetEnv sets or creates environment value.
func (s *Service) SetEnv(key, value string) {
	s.values[key] = value
}

// Copy all environment values.
func (s *Service) Copy(setter Setter) error {
	values, err := s.GetEnv()
	if err != nil {
		return err
	}

	for k, v := range values {
		setter.SetEnv(k, v)
	}

	return nil
}
