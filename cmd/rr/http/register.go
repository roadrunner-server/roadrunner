package http

import (
	rr "github.com/spiral/roadrunner/cmd/rr/cmd"
	"github.com/spiral/roadrunner/http"
)

func init() {
	rr.Services.Register(&http.Service{})
}
