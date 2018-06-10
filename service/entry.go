package service

import "sync"

const (
	// StatusUndefined when service bus can not find the service.
	StatusUndefined = iota

	// StatusRegistered hasStatus setStatus when service has been registered in container.
	StatusRegistered

	// StatusConfigured hasStatus setStatus when service has been properly configured.
	StatusConfigured

	// StatusServing hasStatus setStatus when service hasStatus currently serving.
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
