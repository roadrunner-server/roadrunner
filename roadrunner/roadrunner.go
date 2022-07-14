package roadrunner

import (
	"fmt"

	configImpl "github.com/roadrunner-server/config/v2"
	endure "github.com/roadrunner-server/endure/pkg/container"
	"github.com/roadrunner-server/endure/pkg/fsm"
	"github.com/roadrunner-server/roadrunner/v2/internal/container"
	"github.com/roadrunner-server/roadrunner/v2/internal/meta"
)

const (
	rrPrefix string = "rr"
)

type RR struct {
	container *endure.Endure
	stop      chan struct{}
	Version   string
	BuildTime string
}

// NewRR creates a new RR instance that can then be started or stopped by the caller
func NewRR(cfgFile string, override *[]string, pluginList []interface{}) (*RR, error) {
	// create endure container config
	containerCfg, err := container.NewConfig(cfgFile)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// register config plugin
	err = endureContainer.Register(cfg)
	if err != nil {
		return nil, err
	}

	// register another container plugins
	err = endureContainer.RegisterAll(pluginList...)
	if err != nil {
		return nil, err
	}

	// init container and all services
	err = endureContainer.Init()
	if err != nil {
		return nil, err
	}

	rr := &RR{
		container: endureContainer,
		stop:      make(chan struct{}),
		Version:   meta.Version(),
		BuildTime: meta.BuildTime(),
	}

	return rr, nil
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
		return nil
	}
}

func (rr *RR) CurrentState() fsm.State {
	return rr.container.CurrentState()
}

// Stop stops roadrunner
func (rr *RR) Stop() error {
	rr.stop <- struct{}{}
	return rr.container.Stop()
}

// DefaultPluginsList returns all the plugins that RR can run with and are included by default
func DefaultPluginsList() []interface{} {
	return container.Plugins()
}
