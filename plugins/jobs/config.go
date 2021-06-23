package jobs

import (
	"github.com/spiral/errors"
	poolImpl "github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/plugins/jobs/dispatcher"
	"github.com/spiral/roadrunner/v2/plugins/jobs/pipeline"
	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
)

// Config defines settings for job broker, workers and job-pipeline mapping.
type Config struct {
	// Workers configures roadrunner server and worker busy.
	// Workers *roadrunner.ServerConfig
	poolCfg *poolImpl.Config

	// Dispatch defines where and how to match jobs.
	Dispatch map[string]*structs.Options

	// Pipelines defines mapping between PHP job pipeline and associated job broker.
	Pipelines map[string]*pipeline.Pipeline

	// Consuming specifies names of pipelines to be consumed on service start.
	Consume []string

	// parent config for broken options.
	pipelines pipeline.Pipelines
	route     dispatcher.Dispatcher
}

func (c *Config) InitDefaults() error {
	const op = errors.Op("config_init_defaults")
	var err error
	c.pipelines, err = pipeline.InitPipelines(c.Pipelines)
	if err != nil {
		return errors.E(op, err)
	}

	if c.poolCfg != nil {
		c.poolCfg.InitDefaults()
	}

	return nil
}

// MatchPipeline locates the pipeline associated with the job.
func (c *Config) MatchPipeline(job *structs.Job) (*pipeline.Pipeline, *structs.Options, error) {
	const op = errors.Op("config_match_pipeline")
	opt := c.route.Match(job)

	pipe := ""
	if job.Options != nil {
		pipe = job.Options.Pipeline
	}

	if pipe == "" && opt != nil {
		pipe = opt.Pipeline
	}

	if pipe == "" {
		return nil, nil, errors.E(op, errors.Errorf("unable to locate pipeline for `%s`", job.Job))
	}

	if p := c.pipelines.Get(pipe); p != nil {
		return p, opt, nil
	}

	return nil, nil, errors.E(op, errors.Errorf("undefined pipeline `%s`", pipe))
}
