package jobs

import (
	"runtime"

	poolImpl "github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
)

const (
	// name used to set pipeline name
	pipelineName string = "name"
)

// Config defines settings for job broker, workers and job-pipeline mapping.
type Config struct {
	// NumPollers configures number of priority queue pollers
	// Should be no more than 255
	// Default - num logical cores
	NumPollers uint8 `mapstructure:"num_pollers"`

	// PipelineSize is the limit of a main jobs queue which consume Items from the drivers pipeline
	// Driver pipeline might be much larger than a main jobs queue
	PipelineSize uint64 `mapstructure:"pipeline_size"`

	// Timeout in seconds is the per-push limit to put the job into queue
	Timeout int `mapstructure:"timeout"`

	// Pool configures roadrunner workers pool.
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

	if c.PipelineSize == 0 {
		c.PipelineSize = 1_000_000
	}

	if c.NumPollers == 0 {
		c.NumPollers = uint8(runtime.NumCPU())
	}

	for k := range c.Pipelines {
		// set the pipeline name
		c.Pipelines[k].With(pipelineName, k)
	}

	if c.Timeout == 0 {
		c.Timeout = 10
	}

	c.Pool.InitDefaults()
}
