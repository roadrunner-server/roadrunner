package ws

/*
broadcast:
  ws-us-region-1:
    subscriber: ws
    path: "/ws"

    driver: redis
    address:
      - 6379
    db: 0
*/

// Config represents configuration for the ws plugin
type Config struct {
	// http path for the websocket
	Path string `mapstructure:"Path"`
}

// InitDefault initialize default values for the ws config
func (c *Config) InitDefault() {
	if c.Path == "" {
		c.Path = "/ws"
	}
}
