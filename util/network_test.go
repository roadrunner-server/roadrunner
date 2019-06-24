package util

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestCreateListener(t *testing.T) {
	_, err := CreateListener("unexpected dsn")
	assert.Error(t, err, "Invalid DSN (tcp://:6001, unix://file.sock)")

	_, err = CreateListener("aaa://192.168.0.1")
	assert.Error(t, err, "Invalid Protocol (tcp://:6001, unix://file.sock)")
}

func TestUnixCreateListener(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	l, err := CreateListener("unix://file.sock")
	assert.NoError(t, err)
	l.Close()
}
