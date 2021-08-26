package amqpjobs

// pipeline rabbitmq info
const (
	exchangeKey   string = "exchange"
	exchangeType  string = "exchange_type"
	queue         string = "queue"
	routingKey    string = "routing_key"
	prefetch      string = "prefetch"
	exclusive     string = "exclusive"
	priority      string = "priority"
	multipleAsk   string = "multiple_ask"
	requeueOnFail string = "requeue_on_fail"

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
	Prefetch      int    `mapstructure:"prefetch"`
	Queue         string `mapstructure:"queue"`
	Priority      int64  `mapstructure:"priority"`
	Exchange      string `mapstructure:"exchange"`
	ExchangeType  string `mapstructure:"exchange_type"`
	RoutingKey    string `mapstructure:"routing_key"`
	Exclusive     bool   `mapstructure:"exclusive"`
	MultipleAck   bool   `mapstructure:"multiple_ask"`
	RequeueOnFail bool   `mapstructure:"requeue_on_fail"`
}

func (c *Config) InitDefault() {
	// all options should be in sync with the pipeline defaults in the FromPipeline method
	if c.ExchangeType == "" {
		c.ExchangeType = "direct"
	}

	if c.Exchange == "" {
		c.Exchange = "amqp.default"
	}

	if c.Queue == "" {
		c.Queue = "default"
	}

	if c.Prefetch == 0 {
		c.Prefetch = 10
	}

	if c.Priority == 0 {
		c.Priority = 10
	}
}

func (c *GlobalCfg) InitDefault() {
	if c.Addr == "" {
		c.Addr = "amqp://guest:guest@127.0.0.1:5672/"
	}
}
