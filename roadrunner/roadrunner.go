package roadrunner

import (
	"fmt"

	configImpl "github.com/roadrunner-server/config/v2"
	endure "github.com/roadrunner-server/endure/pkg/container"
	"github.com/roadrunner-server/errors"
	"github.com/roadrunner-server/roadrunner/v2/internal/container"
	"github.com/roadrunner-server/roadrunner/v2/internal/meta"
)

const (
	rrPrefix string = "rr"
)

type RR struct {
	container *endure.Endure
	Version   string
	BuildTime string
}

// NewRR creates a new RR instance that can then be started or stopped by the caller
func NewRR(cfgFile string, override *[]string, pluginList []interface{}) (*RR, error) {
	const op = errors.Op("new_rr")
	// create endure container config
	containerCfg, err := container.NewConfig(cfgFile)
	if err != nil {
		return nil, errors.E(op, err)
	}

	cfg := &configImpl.Plugin{
		Path:    cfgFile,
		Prefix:  rrPrefix,
		Timeout: containerCfg.GracePeriod,
		Flags:   *override,
		Version: meta.Version(),
	}

	// create endure container
	endureContainer, err := container.NewContainer(*containerCfg)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// register config plugin
	if err = endureContainer.Register(cfg); err != nil {
		return nil, errors.E(op, err)
	}

	// register another container plugins
	for i := 0; i < len(pluginList); i++ {
		if err = endureContainer.Register(pluginList[i]); err != nil {
			return nil, errors.E(op, err)
		}
	}

	// init container and all services
	if err = endureContainer.Init(); err != nil {
		return nil, errors.E(op, err)
	}

	rr := &RR{
		container: endureContainer,
		Version:   meta.Version(),
		BuildTime: meta.BuildTime(),
	}

	return rr, nil
}

// Serve starts RR and starts listening for requests.
// This is a blocking call that will return an error if / when one occurs in a plugin
func (rr *RR) Serve() error {
	const op = errors.Op("rr.serve")
	// start serving the graph
	errCh, err := rr.container.Serve()
	if err != nil {
		return errors.E(op, err)
	}

	for {
		select {
		case e := <-errCh:
			rr.Stop()
			return fmt.Errorf("error: %w\nplugin: %s", e.Error, e.VertexID)
		}
	}
}

// Stop stops roadrunner
func (rr *RR) Stop() error {
	if err := rr.container.Stop(); err != nil {
		return fmt.Errorf("error: %w", err)
	}
	return nil
}

// DefaultPluginsList returns all the plugins that RR can run with and are included by default
func DefaultPluginsList() []interface{} {
	return container.Plugins()
}
