package container_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/roadrunner-server/endure/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewContainer(t *testing.T) { // there is no legal way to test container options
	c := endure.New(slog.LevelDebug, endure.Visualize(), endure.GracefulShutdownTimeout(time.Second))
	c2 := endure.New(slog.LevelDebug, endure.Visualize(), endure.GracefulShutdownTimeout(time.Second))

	assert.NotNil(t, c)

	assert.NotNil(t, c2)
}
