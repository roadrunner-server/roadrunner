package container_test

import (
	"testing"
	"time"

	endure "github.com/roadrunner-server/endure/pkg/container"
	"github.com/roadrunner-server/roadrunner/v2/container"
	"github.com/stretchr/testify/assert"
)

func TestNewContainer(t *testing.T) { // there is no legal way to test container options
	c, err := container.NewContainer(container.Config{})
	c2, err2 := container.NewContainer(container.Config{
		GracePeriod: time.Second,
		PrintGraph:  true,
		LogLevel:    endure.WarnLevel,
	})

	assert.NoError(t, err)
	assert.NotNil(t, c)

	assert.NoError(t, err2)
	assert.NotNil(t, c2)
}
