package cmd

import "time"

var (
	// Version - defines build version.
	Version = "local"

	// BuildTime - defined build time.
	BuildTime = time.Now().Format(time.RFC1123)
)
