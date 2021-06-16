package websockets

import (
	"strings"
)

func isOriginAllowed(origin string, cfg *Config) bool {
	if cfg.allowedAll {
		return true
	}

	origin = strings.ToLower(origin)
	// simple case
	origin = strings.ToLower(origin)
	for _, o := range cfg.allowedOrigins {
		if o == origin {
			return true
		}
	}
	// check wildcards
	for _, w := range cfg.allowedWOrigins {
		if w.match(origin) {
			return true
		}
	}

	return false
}
