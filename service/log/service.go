package log

import (
	"github.com/sirupsen/logrus"
	"github.com/spiral/roadrunner/service/rpc"
)

// ID declares the public service name
const ID = "log"

// Service to register the RPC server
type Service struct{}

// Init service.
func (s *Service) Init(log logrus.FieldLogger, r *rpc.Service) (bool, error) {

	if r != nil {
		if err := r.Register("log", &rpcServer{log}); err != nil {
			return false, err
		}
	}

	return true, nil
}
