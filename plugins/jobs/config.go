package jobs

import (
	poolImpl "github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
)

// Config defines settings for job broker, workers and job-pipeline mapping.
type Config struct {
	// Workers configures roadrunner server and worker busy.
	// Workers *roadrunner.ServerConfig
	poolCfg *poolImpl.Config

	// Pipelines defines mapping between PHP job pipeline and associated job broker.
	Pipelines map[string]*pipeline.Pipeline

	// Consuming specifies names of pipelines to be consumed on service start.
	Consume []string
}

func (c *Config) InitDefaults() {
	if c.poolCfg == nil {
		c.poolCfg = &poolImpl.Config{}
	}

	c.poolCfg.InitDefaults()
}
