package lib

import (
	"fmt"
	"runtime/debug"

	configImpl "github.com/roadrunner-server/config/v4"
	"github.com/roadrunner-server/endure/v2"
	"github.com/roadrunner-server/roadrunner/v2023/container"
)

const (
	rrPrefix string = "rr"
	rrModule string = "github.com/roadrunner-server/roadrunner/v2023"
)

type RR struct {
	container *endure.Endure
	stop      chan struct{}
	Version   string
}

// NewRR creates a new RR instance that can then be started or stopped by the caller
func NewRR(cfgFile string, override []string, pluginList []any) (*RR, error) {
	// create endure container config
	containerCfg, err := container.NewConfig(cfgFile)
	if err != nil {
		return nil, err
	}

	cfg := &configImpl.Plugin{
		Path:    cfgFile,
		Prefix:  rrPrefix,
		Timeout: containerCfg.GracePeriod,
		Flags:   override,
		Version: getRRVersion(),
	}

	// create endure container
	endureOptions := []endure.Options{
		endure.GracefulShutdownTimeout(containerCfg.GracePeriod),
	}

	if containerCfg.PrintGraph {
		endureOptions = append(endureOptions, endure.Visualize())
	}

	endureContainer := endure.New(containerCfg.LogLevel, endureOptions...)

	// register another container plugins
	err = endureContainer.RegisterAll(append(pluginList, cfg)...)
	if err != nil {
		return nil, err
	}

	// init container and all services
	err = endureContainer.Init()
	if err != nil {
		return nil, err
	}

	return &RR{
		container: endureContainer,
		stop:      make(chan struct{}, 1),
		Version:   cfg.Version,
	}, nil
}

// Serve starts RR and starts listening for requests.
// This is a blocking call that will return an error if / when one occurs in a plugin
func (rr *RR) Serve() error {
	// start serving the graph
	errCh, err := rr.container.Serve()
	if err != nil {
		return err
	}

	select {
	case e := <-errCh:
		return fmt.Errorf("error: %w\nplugin: %s", e.Error, e.VertexID)
	case <-rr.stop:
		return rr.container.Stop()
	}
}

func (rr *RR) Plugins() []string {
	return rr.container.Plugins()
}

// Stop stops roadrunner
func (rr *RR) Stop() {
	rr.stop <- struct{}{}
}

// DefaultPluginsList returns all the plugins that RR can run with and are included by default
func DefaultPluginsList() []any {
	return container.Plugins()
}

// Tries to find the version info for a given module's path
// empty string if not found
func getRRVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	for i := 0; i < len(bi.Deps); i++ {
		if bi.Deps[i].Path == rrModule {
			return bi.Deps[i].Version
		}
	}

	return ""
}
