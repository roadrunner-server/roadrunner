package checker

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
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
	registry map[string]Checker
	server   *fiber.App
	log      logger.Logger
	cfg      *Config
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

	c.registry = make(map[string]Checker)
	c.log = log
	return nil
}

func (c *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	c.server = fiber.New(fiber.Config{
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
		IdleTimeout:  time.Second * 5,
	})
	c.server.Group("/v1", c.healthHandler)
	c.server.Use(fiberLogger.New())
	c.server.Use("/health", c.healthHandler)

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

// Reset named service.
func (c *Plugin) Status(name string) (Status, error) {
	const op = errors.Op("checker_plugin_status")
	svc, ok := c.registry[name]
	if !ok {
		return Status{}, errors.E(op, errors.Errorf("no such service: %s", name))
	}

	return svc.Status(), nil
}

// CollectTarget collecting services which can provide Status.
func (c *Plugin) CollectTarget(name endure.Named, r Checker) error {
	c.registry[name.Name()] = r
	return nil
}

// Collects declares services to be collected.
func (c *Plugin) Collects() []interface{} {
	return []interface{}{
		c.CollectTarget,
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
		_, _ = ctx.WriteString("No plugins provided in query. Query should be in form of: /v1/health?plugin=plugin1&plugin=plugin2 \n")
		return nil
	}

	failed := false
	// iterate over all provided plugins
	for i := 0; i < len(plugins.Plugins); i++ {
		// check if the plugin exists
		if plugin, ok := c.registry[plugins.Plugins[i]]; ok {
			st := plugin.Status()
			if st.Code >= 500 {
				failed = true
				continue
			} else if st.Code >= 100 && st.Code <= 400 {
				_, _ = ctx.WriteString(fmt.Sprintf(template, plugins.Plugins[i], st.Code))
			}
		} else {
			_, _ = ctx.WriteString(fmt.Sprintf("Service: %s not found", plugins.Plugins[i]))
		}
	}
	if failed {
		ctx.Status(http.StatusInternalServerError)
		return nil
	}

	ctx.Status(http.StatusOK)
	return nil
}
