package kv

import (
	"fmt"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const PluginName string = "kv"

const (
	// driver is the mandatory field which should present in every storage
	driver string = "driver"

	memcached string = "memcached"
	boltdb    string = "boltdb"
	redis     string = "redis"
	memory    string = "memory"
)

// Plugin for the unified storage
type Plugin struct {
	log logger.Logger
	// drivers contains general storage drivers, such as boltdb, memory, memcached, redis.
	drivers map[string]StorageDriver
	// storages contains user-defined storages, such as boltdb-north, memcached-us and so on.
	storages map[string]Storage
	// KV configuration
	cfg       Config
	cfgPlugin config.Configurer
}

func (p *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("kv_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &p.cfg.Data)
	if err != nil {
		return errors.E(op, err)
	}
	p.drivers = make(map[string]StorageDriver, 5)
	p.storages = make(map[string]Storage, 5)
	p.log = log
	p.cfgPlugin = cfg
	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	const op = errors.Op("kv_plugin_serve")
	// key - storage name in the config
	// value - storage
	/*
			For example we can have here 2 storages (but they are not pre-configured)
			for the boltdb and memcached
			We should provide here the actual configs for the all requested storages
				kv:
				  boltdb-south:
				    driver: boltdb
				    dir: "tests/rr-bolt"
				    file: "rr.db"
				    bucket: "rr"
				    permissions: 777
				    ttl: 40s

				  boltdb-north:
					driver: boltdb
					dir: "tests/rr-bolt"
					file: "rr.db"
					bucket: "rr"
					permissions: 777
					ttl: 40s

				  memcached:
				    driver: memcached
				    addr: [ "localhost:11211" ]


		For this config we should have 3 drivers: memory, boltdb and memcached but 4 KVs: default, boltdb-south, boltdb-north and memcached
		when user requests for example boltdb-south, we should provide that particular preconfigured storage
	*/
	for k, v := range p.cfg.Data {
		if _, ok := v.(map[string]interface{})[driver]; !ok {
			errCh <- errors.E(op, errors.Errorf("could not find mandatory driver field in the %s storage", k))
			return errCh
		}

		// config key for the particular sub-driver
		configKey := fmt.Sprintf("%s.%s", PluginName, k)
		// at this point we know, that driver field present in the cofiguration
		switch v.(map[string]interface{})[driver] {
		case memcached:
			if _, ok := p.drivers[memcached]; !ok {
				p.log.Warn("no memcached drivers registered", "registered", p.drivers)
				continue
			}

			storage, err := p.drivers[memcached].Provide(configKey)
			if err != nil {
				errCh <- errors.E(op, err)
				return errCh
			}

			// save the storage
			p.storages[k] = storage

		case boltdb:
			if _, ok := p.drivers[boltdb]; !ok {
				p.log.Warn("no boltdb drivers registered", "registered", p.drivers)
				continue
			}

			storage, err := p.drivers[boltdb].Provide(configKey)
			if err != nil {
				errCh <- errors.E(op, err)
				return errCh
			}

			// save the storage
			p.storages[k] = storage
		case memory:
			if _, ok := p.drivers[memory]; !ok {
				p.log.Warn("no in-memory drivers registered", "registered", p.drivers)
				continue
			}

			storage, err := p.drivers[memory].Provide(configKey)
			if err != nil {
				errCh <- errors.E(op, err)
				return errCh
			}

			// save the storage
			p.storages[k] = storage
		case redis:
			if _, ok := p.drivers[redis]; !ok {
				p.log.Warn("no redis drivers registered", "registered", p.drivers)
				continue
			}

			storage, err := p.drivers[redis].Provide(configKey)
			if err != nil {
				errCh <- errors.E(op, err)
				return errCh
			}

			// save the storage
			p.storages[k] = storage

		default:
			p.log.Error("unknown storage", errors.E(op, errors.Errorf("unknown storage %s", v.(map[string]interface{})[driver])))
		}
	}

	return errCh
}

func (p *Plugin) Stop() error {
	return nil
}

// Collects will get all plugins which implement Storage interface
func (p *Plugin) Collects() []interface{} {
	return []interface{}{
		p.GetAllStorageDrivers,
	}
}

func (p *Plugin) GetAllStorageDrivers(name endure.Named, storage StorageDriver) {
	// save the storage driver
	p.drivers[name.Name()] = storage
}

// RPC returns associated rpc service.
func (p *Plugin) RPC() interface{} {
	return &rpc{srv: p, log: p.log, storages: p.storages}
}

func (p *Plugin) Name() string {
	return PluginName
}
