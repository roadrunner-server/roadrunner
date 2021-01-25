package client

import (
	"fmt"
	"os"
	"sync/atomic"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	rrt "github.com/spiral/roadrunner/v2/plugins/temporal/protocol"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/worker"
)

// PluginName defines public service name.
const PluginName = "temporal"

// indicates that the case size was set
var stickyCacheSet = false

// Plugin implement Temporal contract.
type Plugin struct {
	workerID int32
	cfg      Config
	dc       converter.DataConverter
	log      logger.Logger
	client   client.Client
}

// Temporal define common interface for RoadRunner plugins.
type Temporal interface {
	GetClient() (client.Client, error)
	GetDataConverter() converter.DataConverter
	GetConfig() Config
	GetCodec() rrt.Codec
	CreateWorker(taskQueue string, options worker.Options) (worker.Worker, error)
}

// Config of the temporal client and depended services.
type Config struct {
	Address    string
	Namespace  string
	Activities *pool.Config
	Codec      string
	DebugLevel int `mapstructure:"debug_level"`
	CacheSize  int `mapstructure:"cache_size"`
}

// Init initiates temporal client plugin.
func (p *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	p.log = log
	p.dc = rrt.NewDataConverter(converter.GetDefaultDataConverter())

	return cfg.UnmarshalKey(PluginName, &p.cfg)
}

// GetConfig returns temporal configuration.
func (p *Plugin) GetConfig() Config {
	return p.cfg
}

// GetCodec returns communication codec.
func (p *Plugin) GetCodec() rrt.Codec {
	if p.cfg.Codec == "json" {
		return rrt.NewJSONCodec(rrt.DebugLevel(p.cfg.DebugLevel), p.log)
	}

	// production ready protocol, no debug abilities
	return rrt.NewProtoCodec()
}

// GetDataConverter returns data active data converter.
func (p *Plugin) GetDataConverter() converter.DataConverter {
	return p.dc
}

// Serve starts temporal srv.
func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	var err error

	if stickyCacheSet == false && p.cfg.CacheSize != 0 {
		worker.SetStickyWorkflowCacheSize(p.cfg.CacheSize)
		stickyCacheSet = true
	}

	p.client, err = client.NewClient(client.Options{
		Logger:        p.log,
		HostPort:      p.cfg.Address,
		Namespace:     p.cfg.Namespace,
		DataConverter: p.dc,
	})

	p.log.Debug("Connected to temporal server", "Plugin", p.cfg.Address)

	if err != nil {
		errCh <- errors.E(errors.Op("p connect"), err)
	}

	return errCh
}

// Stop stops temporal srv connection.
func (p *Plugin) Stop() error {
	if p.client != nil {
		p.client.Close()
	}

	return nil
}

// GetClient returns active srv connection.
func (p *Plugin) GetClient() (client.Client, error) {
	return p.client, nil
}

// CreateWorker allocates new temporal worker on an active connection.
func (p *Plugin) CreateWorker(tq string, options worker.Options) (worker.Worker, error) {
	if p.client == nil {
		return nil, errors.E("unable to create worker, invalid temporal p")
	}

	if options.Identity == "" {
		if tq == "" {
			tq = client.DefaultNamespace
		}

		// ensures unique worker IDs
		options.Identity = fmt.Sprintf(
			"%d@%s@%s@%v",
			os.Getpid(),
			getHostName(),
			tq,
			atomic.AddInt32(&p.workerID, 1),
		)
	}

	return worker.New(p.client, tq, options), nil
}

// Name of the service.
func (p *Plugin) Name() string {
	return PluginName
}

func getHostName() string {
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "Unknown"
	}
	return hostName
}
