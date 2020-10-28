package logger

type Config struct {
	Default LoggerConfig

	Suppress bool

	Channels map[string]LoggerConfig
}

type LoggerConfig struct {
	// Level to report messages from.
	Level string
}
