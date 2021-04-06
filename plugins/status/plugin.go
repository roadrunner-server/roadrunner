package status

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const (
	// PluginName declares public plugin name.
	PluginName = "status"
)

type Plugin struct {
	// plugins which needs to be checked just as Status
	statusRegistry map[string]Checker
	// plugins which needs to send Readiness status
	readyRegistry map[string]Readiness
	server        *fiber.App
	log           logger.Logger
	cfg           *Config
}

func (c *Plugin) Init(log logger.Logger, cfg config.Configurer) error {
	const op = errors.Op("checker_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}
	err := cfg.UnmarshalKey(PluginName, &c.cfg)
	if err != nil {
		return errors.E(op, errors.Disabled, err)
	}

	// init defaults for the status plugin
	c.cfg.InitDefaults()

	c.readyRegistry = make(map[string]Readiness)
	c.statusRegistry = make(map[string]Checker)

	c.log = log

	return nil
}

func (c *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	c.server = fiber.New(fiber.Config{
		ReadTimeout:           time.Second * 5,
		WriteTimeout:          time.Second * 5,
		IdleTimeout:           time.Second * 5,
		DisableStartupMessage: true,
	})

	c.server.Use("/health", c.healthHandler)
	c.server.Use("/ready", c.readinessHandler)

	go func() {
		err := c.server.Listen(c.cfg.Address)
		if err != nil {
			errCh <- err
		}
	}()

	return errCh
}

func (c *Plugin) Stop() error {
	const op = errors.Op("checker_plugin_stop")
	err := c.server.Shutdown()
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// status returns a Checker interface implementation
// Reset named service. This is not an Status interface implementation
func (c *Plugin) status(name string) (Status, error) {
	const op = errors.Op("checker_plugin_status")
	svc, ok := c.statusRegistry[name]
	if !ok {
		return Status{}, errors.E(op, errors.Errorf("no such service: %s", name))
	}

	return svc.Status(), nil
}

// ready used to provide a readiness check for the plugin
func (c *Plugin) ready(name string) (Status, error) {
	const op = errors.Op("checker_plugin_ready")
	svc, ok := c.readyRegistry[name]
	if !ok {
		return Status{}, errors.E(op, errors.Errorf("no such service: %s", name))
	}

	return svc.Ready(), nil
}

// CollectCheckerImpls collects services which can provide Status.
func (c *Plugin) CollectCheckerImpls(name endure.Named, r Checker) error {
	c.statusRegistry[name.Name()] = r
	return nil
}

// CollectReadinessImpls collects services which can provide Readiness check.
func (c *Plugin) CollectReadinessImpls(name endure.Named, r Readiness) error {
	c.readyRegistry[name.Name()] = r
	return nil
}

// Collects declares services to be collected.
func (c *Plugin) Collects() []interface{} {
	return []interface{}{
		c.CollectReadinessImpls,
		c.CollectCheckerImpls,
	}
}

// Name of the service.
func (c *Plugin) Name() string {
	return PluginName
}

// RPCService returns associated rpc service.
func (c *Plugin) RPC() interface{} {
	return &rpc{srv: c, log: c.log}
}

type Plugins struct {
	Plugins []string `query:"plugin"`
}

const template string = "Service: %s: Status: %d\n"

func (c *Plugin) healthHandler(ctx *fiber.Ctx) error {
	const op = errors.Op("checker_plugin_health_handler")
	plugins := &Plugins{}
	err := ctx.QueryParser(plugins)
	if err != nil {
		return errors.E(op, err)
	}

	if len(plugins.Plugins) == 0 {
		ctx.Status(http.StatusOK)
		_, _ = ctx.WriteString("No plugins provided in query. Query should be in form of: health?plugin=plugin1&plugin=plugin2 \n")
		return nil
	}

	// iterate over all provided plugins
	for i := 0; i < len(plugins.Plugins); i++ {
		// check if the plugin exists
		if plugin, ok := c.statusRegistry[plugins.Plugins[i]]; ok {
			st := plugin.Status()
			if st.Code >= 500 {
				// if there is 500 or 503 status code return immediately
				ctx.Status(c.cfg.UnavailableStatusCode)
				return nil
			} else if st.Code >= 100 && st.Code <= 400 {
				_, _ = ctx.WriteString(fmt.Sprintf(template, plugins.Plugins[i], st.Code))
			}
		} else {
			_, _ = ctx.WriteString(fmt.Sprintf("Service: %s not found", plugins.Plugins[i]))
		}
	}

	ctx.Status(http.StatusOK)
	return nil
}

// readinessHandler return 200OK if all plugins are ready to serve
// if one of the plugins return status from the 5xx range, the status for all query will be 503
func (c *Plugin) readinessHandler(ctx *fiber.Ctx) error {
	const op = errors.Op("checker_plugin_readiness_handler")
	plugins := &Plugins{}
	err := ctx.QueryParser(plugins)
	if err != nil {
		return errors.E(op, err)
	}

	if len(plugins.Plugins) == 0 {
		ctx.Status(http.StatusOK)
		_, _ = ctx.WriteString("No plugins provided in query. Query should be in form of: ready?plugin=plugin1&plugin=plugin2 \n")
		return nil
	}

	// iterate over all provided plugins
	for i := 0; i < len(plugins.Plugins); i++ {
		// check if the plugin exists
		if plugin, ok := c.readyRegistry[plugins.Plugins[i]]; ok {
			st := plugin.Ready()
			if st.Code >= 500 {
				// if there is 500 or 503 status code return immediately
				ctx.Status(c.cfg.UnavailableStatusCode)
				return nil
			} else if st.Code >= 100 && st.Code <= 400 {
				_, _ = ctx.WriteString(fmt.Sprintf(template, plugins.Plugins[i], st.Code))
			}
		} else {
			_, _ = ctx.WriteString(fmt.Sprintf("Service: %s not found", plugins.Plugins[i]))
		}
	}

	ctx.Status(http.StatusOK)
	return nil
}
