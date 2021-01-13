// +build windows

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateListener(t *testing.T) {
	_, err := CreateListener("unexpected dsn")
	assert.Error(t, err, "Invalid DSN (tcp://:6001, unix://file.sock)")

	_, err = CreateListener("aaa://192.168.0.1")
	assert.Error(t, err, "Invalid Protocol (tcp://:6001, unix://file.sock)")
}
