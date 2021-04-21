package kv

type Config struct {
	Data map[string]interface{} `mapstructure:"kv"`
}
