package checker

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/spiral/endure"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/interfaces/log"
	"github.com/spiral/roadrunner/v2/interfaces/status"
	"github.com/spiral/roadrunner/v2/plugins/config"
)

const (
	// PluginName declares public plugin name.
	PluginName = "status"
)

type Plugin struct {
	registry map[string]status.Checker
	server   *fiber.App
	log      log.Logger
	cfg      *Config
}

func (c *Plugin) Init(log log.Logger, cfg config.Configurer) error {
	const op = errors.Op("status plugin init")
	err := cfg.UnmarshalKey(PluginName, &c.cfg)
	if err != nil {
		return errors.E(op, err)
	}
	c.registry = make(map[string]status.Checker)
	c.log = log
	return nil
}

// localhost:88294/status/all
func (c *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	c.server = fiber.New()
	c.server.Group("/v1", c.healthHandler)
	c.server.Use(logger.New())
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
	return c.server.Shutdown()
}

// Reset named service.
func (c *Plugin) Status(name string) (status.Status, error) {
	const op = errors.Op("get status")
	svc, ok := c.registry[name]
	if !ok {
		return status.Status{}, errors.E(op, errors.Errorf("no such service: %s", name))
	}

	return svc.Status(), nil
}

// CollectTarget collecting services which can provide Status.
func (c *Plugin) CollectTarget(name endure.Named, r status.Checker) error {
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
	const op = errors.Op("health_handler")
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
