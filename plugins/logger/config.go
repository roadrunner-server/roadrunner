package logger

type Config struct {
	Squash   bool
	Channels map[string]LoggerConfig
}

type LoggerConfig struct {
}
