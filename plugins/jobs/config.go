package jobs

import (
	poolImpl "github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
)

// Config defines settings for job broker, workers and job-pipeline mapping.
type Config struct {
	// Workers configures roadrunner server and worker busy.
	// Workers *roadrunner.ServerConfig
	Pool *poolImpl.Config `mapstructure:"Pool"`

	// Pipelines defines mapping between PHP job pipeline and associated job broker.
	Pipelines map[string]*pipeline.Pipeline `mapstructure:"pipelines"`

	// Consuming specifies names of pipelines to be consumed on service start.
	Consume []string `mapstructure:"consume"`
}

func (c *Config) InitDefaults() {
	if c.Pool == nil {
		c.Pool = &poolImpl.Config{}
	}

	c.Pool.InitDefaults()
}
