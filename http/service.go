package http

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/spiral/roadrunner/service"
	"net/http"
	"github.com/spiral/roadrunner"
)

type Service struct {
	cfg  *serviceConfig
	http *http.Server
	srv  *Server
}

func (s *Service) Name() string {
	return "http"
}

func (s *Service) Configure(cfg service.Config) (bool, error) {
	config := &serviceConfig{}
	if err := cfg.Unmarshal(config); err != nil {
		return false, err
	}

	if !config.Enabled {
		return false, nil
	}

	if err := config.Valid(); err != nil {
		return false, err
	}

	s.cfg = config
	return true, nil
}

func (s *Service) RPC() interface{} {
	return &rpcServer{s}
}

func (s *Service) Serve() error {
	logrus.Debugf("http: started")
	defer logrus.Debugf("http: stopped")

	rr, term, err := s.cfg.Pool.NewServer()
	if err != nil {
		return err
	}
	defer term()

	//todo: remove
	rr.Observe(func(event int, ctx interface{}) {
		switch event {
		case roadrunner.EventPoolError:
			logrus.Error(ctx)
		case roadrunner.EventWorkerError:
			logrus.Errorf("%s: %s", ctx.(roadrunner.WorkerError).Worker, ctx.(roadrunner.WorkerError).Error())
		}
	})

	s.srv = NewServer(s.cfg.httpConfig(), rr)
	s.http = &http.Server{
		Addr:    s.cfg.httpAddr(),
		Handler: s.srv,
	}

	if err := s.http.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func (s *Service) Stop() error {
	return s.http.Shutdown(context.Background())
}
