// Package helpers provides RPC helper functions for end-to-end tests.
// Each helper returns a func(t *testing.T) suitable for use with t.Run(),
// performing operations like pushing jobs, pausing pipelines, and collecting
// statistics via Goridge RPC.
package helpers
