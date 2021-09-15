package grpc

import (
	"time"

	"github.com/spiral/roadrunner/v2/pkg/pool"
)

type Config struct {
	Listen string `mapstructure:"listen"`
	Proto  string `mapstructure:"proto"`

	TLS *TLS

	// Env is environment variables passed to the http pool
	Env map[string]string

	GrpcPool              pool.Config
	MaxSendMsgSize        int64         `mapstructure:"max_send_msg_size"`
	MaxRecvMsgSize        int64         `mapstructure:"max_recv_msg_size"`
	MaxConnectionIdle     time.Duration `mapstructure:"max_connection_idle"`
	MaxConnectionAge      time.Duration `mapstructure:"max_connection_age"`
	MaxConnectionAgeGrace time.Duration `mapstructure:"max_connection_age_grace"`
	MaxConcurrentStreams  int64         `mapstructure:"max_concurrent_streams"`
	PingTime              time.Duration `mapstructure:"ping_time"`
	Timeout               time.Duration `mapstructure:"timeout"`
}

type TLS struct {
	Key    string
	Cert   string
	RootCA string
}

func (c *Config) InitDefaults() {

}

func (c *Config) EnableTLS() bool {
	if c.TLS != nil {
		return (c.TLS.RootCA != "" && c.TLS.Key != "" && c.TLS.Cert != "") || (c.TLS.Key != "" && c.TLS.Cert != "")
	}

	return false
}
