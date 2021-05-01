package protocols

import (
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/protocols/fcgi"
	"github.com/spiral/roadrunner/v2/plugins/protocols/http"
	"github.com/spiral/roadrunner/v2/plugins/protocols/https"
)

// Plugin combines all protocols under one workers poll
type Plugin struct {
}

// Init combines all available protocols at the moment, http, https, fcgi
func (p *Plugin) Init(log logger.Logger, h *http.Plugin, hs *https.Plugin, fc *fcgi.Plugin) error {
	return nil
}
