package watcher

import "github.com/spiral/roadrunner"

// Server defines the ability to get access to underlying roadrunner server for watching capabilities.
type Server interface {
	// Server must return associated roadrunner serve.
	Server() *roadrunner.Server
}
