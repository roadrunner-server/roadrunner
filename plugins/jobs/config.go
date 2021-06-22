package jobs

import (
	"github.com/spiral/errors"
	poolImpl "github.com/spiral/roadrunner/v2/pkg/pool"
)

// Config defines settings for job broker, workers and job-pipeline mapping.
type Config struct {
	// Workers configures roadrunner server and worker busy.
	// Workers *roadrunner.ServerConfig
	poolCfg poolImpl.Config

	// Dispatch defines where and how to match jobs.
	Dispatch map[string]*Options

	// Pipelines defines mapping between PHP job pipeline and associated job broker.
	Pipelines map[string]*Pipeline

	// Consuming specifies names of pipelines to be consumed on service start.
	Consume []string

	// parent config for broken options.
	pipelines Pipelines
	route     Dispatcher
}

func (c *Config) InitDefaults() error {
	const op = errors.Op("config_init_defaults")
	var err error
	c.pipelines, err = initPipelines(c.Pipelines)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// MatchPipeline locates the pipeline associated with the job.
func (c *Config) MatchPipeline(job *Job) (*Pipeline, *Options, error) {
	const op = errors.Op("config_match_pipeline")
	opt := c.route.match(job)

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
