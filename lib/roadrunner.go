package lib

import (
	"fmt"
	"runtime/debug"

	configImpl "github.com/roadrunner-server/config/v3"
	endure "github.com/roadrunner-server/endure/pkg/container"
	"github.com/roadrunner-server/endure/pkg/fsm"
	"github.com/roadrunner-server/roadrunner/v2/container"
)

const (
	rrPrefix string = "rr"
	rrModule string = "github.com/roadrunner-server/roadrunner/v2"
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
	endureContainer, err := container.NewContainer(*containerCfg)
	if err != nil {
		return nil, err
	}

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

func (rr *RR) CurrentState() fsm.State {
	return rr.container.CurrentState()
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
