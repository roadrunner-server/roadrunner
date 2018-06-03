package service

type Service interface {
	Name() string
	Configure(cfg Config) (bool, error)
	RPC() interface{}
	Serve() error
	Stop() error
}
