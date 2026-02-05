// Package rpc provides an internal RPC client for CLI-to-server communication.
// It handles configuration loading with environment variable substitution,
// flag override parsing, config file includes, and network dialing via the
// Goridge protocol. This package is for internal use only and should be kept
// in sync with the RPC plugin.
package rpc
