package websockets


// Config defines the websocket service configuration.
type Config struct {
	// Path defines on this URL the middleware must be activated. Same path must
	// be handled by underlying application kernel to authorize the consumption.
	Path string

	// NoOrigin disables origin check, only for debug.
	NoOrigin bool
}

// Hydrate reads the configuration values from the source configuration.
//func (c *Config) Hydrate(cfg service.Config) error {
//	if err := cfg.Unmarshal(c); err != nil {
//		return err
//	}
//
//	return nil
//}
