package broadcast

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
	// Sections represent particular broadcast plugin section
	Sections map[string]interface{} `mapstructure:"sections"`
}

func (c *Config) InitDefaults() {

}

func (c *Config) Valid() error {
	return nil
}
