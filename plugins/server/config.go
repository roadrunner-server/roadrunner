package server

import (
	"time"
)

// All config (.rr.yaml)
// For other section use pointer to distinguish between `empty` and `not present`
type Config struct {
	// Server config section
	Server struct {
		// Command to run as application.
		Command string `mapstructure:"command"`
		// User to run application under.
		User string `mapstructure:"user"`
		// Group to run application under.
		Group string `mapstructure:"group"`
		// Env represents application environment.
		Env Env `mapstructure:"env"`
		// Relay defines connection method and factory to be used to connect to workers:
		// "pipes", "tcp://:6001", "unix://rr.sock"
		// This config section must not change on re-configuration.
		Relay string `mapstructure:"relay"`
		// RelayTimeout defines for how long socket factory will be waiting for worker connection. This config section
		// must not change on re-configuration. Defaults to 60s.
		RelayTimeout time.Duration `mapstructure:"relay_timeout"`
	} `mapstructure:"server"`

	RPC *struct {
		Listen string `mapstructure:"listen"`
	} `mapstructure:"rpc"`
	Logs *struct {
		Mode  string `mapstructure:"mode"`
		Level string `mapstructure:"level"`
	} `mapstructure:"logs"`
	HTTP *struct {
		Address        string   `mapstructure:"address"`
		MaxRequestSize int      `mapstructure:"max_request_size"`
		Middleware     []string `mapstructure:"middleware"`
		Uploads        struct {
			Forbid []string `mapstructure:"forbid"`
		} `mapstructure:"uploads"`
		TrustedSubnets []string `mapstructure:"trusted_subnets"`
		Pool           struct {
			NumWorkers      int    `mapstructure:"num_workers"`
			MaxJobs         int    `mapstructure:"max_jobs"`
			AllocateTimeout string `mapstructure:"allocate_timeout"`
			DestroyTimeout  string `mapstructure:"destroy_timeout"`
			Supervisor      struct {
				WatchTick       int `mapstructure:"watch_tick"`
				TTL             int `mapstructure:"ttl"`
				IdleTTL         int `mapstructure:"idle_ttl"`
				ExecTTL         int `mapstructure:"exec_ttl"`
				MaxWorkerMemory int `mapstructure:"max_worker_memory"`
			} `mapstructure:"supervisor"`
		} `mapstructure:"pool"`
		Ssl struct {
			Port     int    `mapstructure:"port"`
			Redirect bool   `mapstructure:"redirect"`
			Cert     string `mapstructure:"cert"`
			Key      string `mapstructure:"key"`
		} `mapstructure:"ssl"`
		Fcgi struct {
			Address string `mapstructure:"address"`
		} `mapstructure:"fcgi"`
		HTTP2 struct {
			Enabled              bool `mapstructure:"enabled"`
			H2C                  bool `mapstructure:"h2c"`
			MaxConcurrentStreams int  `mapstructure:"max_concurrent_streams"`
		} `mapstructure:"http2"`
	} `mapstructure:"http"`
	Redis *struct {
		Addrs            []string `mapstructure:"addrs"`
		MasterName       string   `mapstructure:"master_name"`
		Username         string   `mapstructure:"username"`
		Password         string   `mapstructure:"password"`
		DB               int      `mapstructure:"db"`
		SentinelPassword string   `mapstructure:"sentinel_password"`
		RouteByLatency   bool     `mapstructure:"route_by_latency"`
		RouteRandomly    bool     `mapstructure:"route_randomly"`
		DialTimeout      int      `mapstructure:"dial_timeout"`
		MaxRetries       int      `mapstructure:"max_retries"`
		MinRetryBackoff  int      `mapstructure:"min_retry_backoff"`
		MaxRetryBackoff  int      `mapstructure:"max_retry_backoff"`
		PoolSize         int      `mapstructure:"pool_size"`
		MinIdleConns     int      `mapstructure:"min_idle_conns"`
		MaxConnAge       int      `mapstructure:"max_conn_age"`
		ReadTimeout      int      `mapstructure:"read_timeout"`
		WriteTimeout     int      `mapstructure:"write_timeout"`
		PoolTimeout      int      `mapstructure:"pool_timeout"`
		IdleTimeout      int      `mapstructure:"idle_timeout"`
		IdleCheckFreq    int      `mapstructure:"idle_check_freq"`
		ReadOnly         bool     `mapstructure:"read_only"`
	} `mapstructure:"redis"`
	Boltdb *struct {
		Dir         string `mapstructure:"dir"`
		File        string `mapstructure:"file"`
		Bucket      string `mapstructure:"bucket"`
		Permissions int    `mapstructure:"permissions"`
		TTL         int    `mapstructure:"TTL"`
	} `mapstructure:"boltdb"`
	Memcached *struct {
		Addr []string `mapstructure:"addr"`
	} `mapstructure:"memcached"`
	Memory *struct {
		Enabled  bool `mapstructure:"enabled"`
		Interval int  `mapstructure:"interval"`
	} `mapstructure:"memory"`
	Metrics *struct {
		Address string `mapstructure:"address"`
		Collect struct {
			AppMetric struct {
				Type       string    `mapstructure:"type"`
				Help       string    `mapstructure:"help"`
				Labels     []string  `mapstructure:"labels"`
				Buckets    []float64 `mapstructure:"buckets"`
				Objectives []struct {
					Num2 float64 `mapstructure:"2,omitempty"`
					One4 float64 `mapstructure:"1.4,omitempty"`
				} `mapstructure:"objectives"`
			} `mapstructure:"app_metric"`
		} `mapstructure:"collect"`
	} `mapstructure:"metrics"`
	Reload *struct {
		Interval string   `mapstructure:"interval"`
		Patterns []string `mapstructure:"patterns"`
		Services struct {
			HTTP struct {
				Recursive bool     `mapstructure:"recursive"`
				Ignore    []string `mapstructure:"ignore"`
				Patterns  []string `mapstructure:"patterns"`
				Dirs      []string `mapstructure:"dirs"`
			} `mapstructure:"http"`
		} `mapstructure:"services"`
	} `mapstructure:"reload"`
}

// InitDefaults for the server config
func (cfg *Config) InitDefaults() {
	if cfg.Server.Relay == "" {
		cfg.Server.Relay = "pipes"
	}

	if cfg.Server.RelayTimeout == 0 {
		cfg.Server.RelayTimeout = time.Second * 60
	}
}
