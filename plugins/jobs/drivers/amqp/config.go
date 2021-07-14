package amqp

// pipeline rabbitmq info
const (
	exchangeKey  string = "exchange"
	exchangeType string = "exchange-type"
	queue        string = "queue"
	routingKey   string = "routing-key"
	prefetch     string = "prefetch"
	exclusive    string = "exclusive"
	priority     string = "priority"

	dlx           string = "x-dead-letter-exchange"
	dlxRoutingKey string = "x-dead-letter-routing-key"
	dlxTTL        string = "x-message-ttl"
	dlxExpires    string = "x-expires"

	contentType string = "application/octet-stream"
)

type GlobalCfg struct {
	Addr string `mapstructure:"addr"`
}

// Config is used to parse pipeline configuration
type Config struct {
	PrefetchCount int    `mapstructure:"pipeline_size"`
	Queue         string `mapstructure:"queue"`
	Priority      int64  `mapstructure:"priority"`
	Exchange      string `mapstructure:"exchange"`
	ExchangeType  string `mapstructure:"exchange_type"`
	RoutingKey    string `mapstructure:"routing_key"`
	Exclusive     bool   `mapstructure:"exclusive"`
}

func (c *Config) InitDefault() {
	if c.ExchangeType == "" {
		c.ExchangeType = "direct"
	}

	if c.Exchange == "" {
		c.Exchange = "default"
	}

	if c.PrefetchCount == 0 {
		c.PrefetchCount = 100
	}

	if c.Priority == 0 {
		c.Priority = 10
	}
}

func (c *GlobalCfg) InitDefault() {
	if c.Addr == "" {
		c.Addr = "amqp://guest:guest@localhost:5672/"
	}
}
