package container_test

import (
	"testing"
	"time"

	"github.com/spiral/roadrunner-binary/v2/internal/container"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/stretchr/testify/assert"
)

func TestNewContainer(t *testing.T) { // there is no legal way to test container options
	c, err := container.NewContainer(container.Config{})
	c2, err2 := container.NewContainer(container.Config{
		GracePeriod: time.Second,
		PrintGraph:  true,
		RetryOnFail: true,
		LogLevel:    endure.WarnLevel,
	})

	assert.NoError(t, err)
	assert.NotNil(t, c)

	assert.NoError(t, err2)
	assert.NotNil(t, c2)
}
