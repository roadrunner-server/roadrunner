package service

type Config interface {
	Get(key string) Config
	Unmarshal(out interface{}) error
}
