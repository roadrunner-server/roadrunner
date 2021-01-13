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
		Command string `yaml:"command"`
		// User to run application under.
		User string `yaml:"user"`
		// Group to run application under.
		Group string `yaml:"group"`
		// Env represents application environment.
		Env Env `yaml:"env"`
		// Relay defines connection method and factory to be used to connect to workers:
		// "pipes", "tcp://:6001", "unix://rr.sock"
		// This config section must not change on re-configuration.
		Relay string `yaml:"relay"`
		// RelayTimeout defines for how long socket factory will be waiting for worker connection. This config section
		// must not change on re-configuration. Defaults to 60s.
		RelayTimeout time.Duration `yaml:"relayTimeout"`
	} `yaml:"server"`

	RPC *struct {
		Listen string `yaml:"listen"`
	} `yaml:"rpc"`
	Logs *struct {
		Mode  string `yaml:"mode"`
		Level string `yaml:"level"`
	} `yaml:"logs"`
	HTTP *struct {
		Address        string   `yaml:"address"`
		MaxRequestSize int      `yaml:"max_request_size"`
		Middleware     []string `yaml:"middleware"`
		Uploads        struct {
			Forbid []string `yaml:"forbid"`
		} `yaml:"uploads"`
		TrustedSubnets []string `yaml:"trusted_subnets"`
		Pool           struct {
			NumWorkers      int    `yaml:"num_workers"`
			MaxJobs         int    `yaml:"max_jobs"`
			AllocateTimeout string `yaml:"allocate_timeout"`
			DestroyTimeout  string `yaml:"destroy_timeout"`
			Supervisor      struct {
				WatchTick       int `yaml:"watch_tick"`
				TTL             int `yaml:"ttl"`
				IdleTTL         int `yaml:"idle_ttl"`
				ExecTTL         int `yaml:"exec_ttl"`
				MaxWorkerMemory int `yaml:"max_worker_memory"`
			} `yaml:"supervisor"`
		} `yaml:"pool"`
		Ssl struct {
			Port     int    `yaml:"port"`
			Redirect bool   `yaml:"redirect"`
			Cert     string `yaml:"cert"`
			Key      string `yaml:"key"`
		} `yaml:"ssl"`
		Fcgi struct {
			Address string `yaml:"address"`
		} `yaml:"fcgi"`
		HTTP2 struct {
			Enabled              bool `yaml:"enabled"`
			H2C                  bool `yaml:"h2c"`
			MaxConcurrentStreams int  `yaml:"max_concurrent_streams"`
		} `yaml:"http2"`
	} `yaml:"http"`
	Redis *struct {
		Addrs            []string `yaml:"addrs"`
		MasterName       string   `yaml:"master_name"`
		Username         string   `yaml:"username"`
		Password         string   `yaml:"password"`
		DB               int      `yaml:"db"`
		SentinelPassword string   `yaml:"sentinel_password"`
		RouteByLatency   bool     `yaml:"route_by_latency"`
		RouteRandomly    bool     `yaml:"route_randomly"`
		DialTimeout      int      `yaml:"dial_timeout"`
		MaxRetries       int      `yaml:"max_retries"`
		MinRetryBackoff  int      `yaml:"min_retry_backoff"`
		MaxRetryBackoff  int      `yaml:"max_retry_backoff"`
		PoolSize         int      `yaml:"pool_size"`
		MinIdleConns     int      `yaml:"min_idle_conns"`
		MaxConnAge       int      `yaml:"max_conn_age"`
		ReadTimeout      int      `yaml:"read_timeout"`
		WriteTimeout     int      `yaml:"write_timeout"`
		PoolTimeout      int      `yaml:"pool_timeout"`
		IdleTimeout      int      `yaml:"idle_timeout"`
		IdleCheckFreq    int      `yaml:"idle_check_freq"`
		ReadOnly         bool     `yaml:"read_only"`
	} `yaml:"redis"`
	Boltdb *struct {
		Dir         string `yaml:"dir"`
		File        string `yaml:"file"`
		Bucket      string `yaml:"bucket"`
		Permissions int    `yaml:"permissions"`
		TTL         int    `yaml:"TTL"`
	} `yaml:"boltdb"`
	Memcached *struct {
		Addr []string `yaml:"addr"`
	} `yaml:"memcached"`
	Memory *struct {
		Enabled  bool `yaml:"enabled"`
		Interval int  `yaml:"interval"`
	} `yaml:"memory"`
	Metrics *struct {
		Address string `yaml:"address"`
		Collect struct {
			AppMetric struct {
				Type       string    `yaml:"type"`
				Help       string    `yaml:"help"`
				Labels     []string  `yaml:"labels"`
				Buckets    []float64 `yaml:"buckets"`
				Objectives []struct {
					Num2 float64 `yaml:"2,omitempty"`
					One4 float64 `yaml:"1.4,omitempty"`
				} `yaml:"objectives"`
			} `yaml:"app_metric"`
		} `yaml:"collect"`
	} `yaml:"metrics"`
	Reload *struct {
		Interval string   `yaml:"interval"`
		Patterns []string `yaml:"patterns"`
		Services struct {
			HTTP struct {
				Recursive bool     `yaml:"recursive"`
				Ignore    []string `yaml:"ignore"`
				Patterns  []string `yaml:"patterns"`
				Dirs      []string `yaml:"dirs"`
			} `yaml:"http"`
		} `yaml:"services"`
	} `yaml:"reload"`
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
