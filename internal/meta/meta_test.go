package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	for give, want := range map[string]string{
		// without changes
		"vvv":     "vvv",
		"victory": "victory",
		"voodoo":  "voodoo",
		"foo":     "foo",
		"0.0.0":   "0.0.0",
		"v":       "v",
		"V":       "V",

		// "v" prefix removal
		"v0.0.0": "0.0.0",
		"V0.0.0": "0.0.0",
		"v1":     "1",
		"V1":     "1",

		// with spaces
		" 0.0.0":  "0.0.0",
		"v0.0.0 ": "0.0.0",
		" V0.0.0": "0.0.0",
		"v1 ":     "1",
		" V1":     "1",
		"v ":      "v",
	} {
		version = give

		assert.Equal(t, want, Version())
	}
}

func TestBuildTime(t *testing.T) {
	for give, want := range map[string]string{
		"development":              "development",
		"2021-03-26T13:50:31+0500": "2021-03-26T13:50:31+0500",
	} {
		buildTime = give

		assert.Equal(t, want, BuildTime())
	}
}
