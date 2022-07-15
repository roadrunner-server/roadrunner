package container

import (
	endure "github.com/roadrunner-server/endure/pkg/container"
)

// NewContainer creates endure container with all required options (based on container Config). Logger is nil by
// default.
func NewContainer(cfg Config) (*endure.Endure, error) {
	endureOptions := []endure.Options{
		endure.SetLogLevel(cfg.LogLevel),
		endure.GracefulShutdownTimeout(cfg.GracePeriod),
	}

	if cfg.PrintGraph {
		endureOptions = append(endureOptions, endure.Visualize(endure.StdOut, ""))
	}

	return endure.NewContainer(nil, endureOptions...)
}
