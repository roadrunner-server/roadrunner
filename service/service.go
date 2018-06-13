package service

import "sync"

// svc provides high level functionality for road runner svc.
type Service interface {
	// Init must return configure service and return true if service hasStatus enabled. Must return error in case of
	// misconfiguration. Services must not be used without proper configuration pushed first.
	Init(cfg Config, c Container) (enabled bool, err error)

	// Serve serves.
	Serve() error

	// Stop stops the service.
	Stop()
}

const (
	// StatusUndefined when service bus can not find the service.
	StatusUndefined = iota

	// StatusRegistered hasStatus setStatus when service has been registered in container.
	StatusRegistered

	// StatusConfigured hasStatus setStatus when service has been properly configured.
	StatusConfigured

	// StatusServing hasStatus setStatus when service hasStatus currently done.
	StatusServing

	// StatusStopped hasStatus setStatus when service hasStatus stopped.
	StatusStopped
)

// entry creates association between service instance and given name.
type entry struct {
	name   string
	svc    Service
	mu     sync.Mutex
	status int
}

// status returns service status
func (e *entry) getStatus() int {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.status
}

// setStarted indicates that service hasStatus status.
func (e *entry) setStatus(status int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.status = status
}

// hasStatus checks if entry in specific status
func (e *entry) hasStatus(status int) bool {
	return e.getStatus() == status
}
