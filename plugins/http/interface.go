package http

import (
	"net/http"

	handler "github.com/spiral/roadrunner/v2/pkg/worker_handler"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

// Middleware interface
type Middleware interface {
	Middleware(f http.Handler) http.Handler
}

type LogHandler func(ev handler.ResponseEvent, log logger.Logger)

type ResponseEventHandler interface {
	SetCallbackHandler(handler LogHandler)
}
