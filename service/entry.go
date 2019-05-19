package service

import (
	"sync"
)

const (
	// StatusUndefined when service bus can not find the service.
	StatusUndefined = iota

	// StatusInactive when service has been registered in container.
	StatusInactive

	// StatusOK when service has been properly configured.
	StatusOK

	// StatusServing when service is currently done.
	StatusServing

	// StatusStopping when service is currently stopping.
	StatusStopping

	// StatusStopped when service being stopped.
	StatusStopped
)

// entry creates association between service instance and given name.
type entry struct {
	name   string
	svc    interface{}
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

// canServe returns true is service can serve.
func (e *entry) canServe() bool {
	_, ok := e.svc.(Service)
	return ok
}
