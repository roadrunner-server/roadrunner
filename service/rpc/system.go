package rpc

import "github.com/spiral/roadrunner/service"

// systemService service controls rr server.
type systemService struct {
	c service.Container
}

// Detach the underlying c.
func (s *systemService) Stop(stop bool, r *string) error {
	if stop {
		s.c.Stop()
	}
	*r = "OK"

	return nil
}
