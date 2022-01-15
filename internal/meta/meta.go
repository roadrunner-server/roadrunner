package meta

import "strings"

// next variables will be set during compilation (do NOT rename them).
var (
	version   = "local"
	buildTime = "development" //nolint:gochecknoglobals
)

// Version returns version value (without `v` prefix).
func Version() string {
	v := strings.TrimSpace(version)

	if len(v) > 1 && ((v[0] == 'v' || v[0] == 'V') && (v[1] >= '0' && v[1] <= '9')) {
		return v[1:]
	}

	return v
}

// BuildTime returns application building time.
func BuildTime() string { return buildTime }
