package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_NumWorkers(t *testing.T) {
	cfg := Config{
		AllocateTimeout: time.Second,
		DestroyTimeout:  time.Second * 10,
	}
	err := cfg.Valid()

	assert.NotNil(t, err)
	assert.Equal(t, "cfg.NumWorkers must be set", err.Error())
}

func Test_AllocateTimeout(t *testing.T) {
	cfg := Config{
		NumWorkers:     10,
		DestroyTimeout: time.Second * 10,
	}
	err := cfg.Valid()

	assert.NotNil(t, err)
	assert.Equal(t, "cfg.AllocateTimeout must be set", err.Error())
}

func Test_DestroyTimeout(t *testing.T) {
	cfg := Config{
		NumWorkers:      10,
		AllocateTimeout: time.Second,
	}
	err := cfg.Valid()

	assert.NotNil(t, err)
	assert.Equal(t, "cfg.DestroyTimeout must be set", err.Error())
}
