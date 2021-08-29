package http

import (
	handler "github.com/spiral/roadrunner/v2/pkg/worker_handler"
	"github.com/spiral/roadrunner/v2/plugins/http"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

type PluginLogging struct {
}

func (p1 *PluginLogging) Init(reHandler http.ResponseEventHandler) error {
	reHandler.SetCallbackHandler(func(ev handler.ResponseEvent, logger logger.Logger) {
		logger.Debug("test of custom logging", "url", ev.Request.URI)
	})

	return nil
}

func (p1 *PluginLogging) Name() string {
	return "http_test.plugin_logging"
}
