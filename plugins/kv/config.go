package kv

// Config represents general storage configuration with keys as the user defined kv-names and values as the constructors
type Config struct {
	Data map[string]interface{} `mapstructure:"kv"`
}
