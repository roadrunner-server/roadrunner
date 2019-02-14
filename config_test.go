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
	assert.Equal(t, "pool.NumWorkers must be set", err.Error())
}

func Test_NumWorkers_Default(t *testing.T) {
	cfg := Config{
		AllocateTimeout: time.Second,
		DestroyTimeout:  time.Second * 10,
	}

	assert.NoError(t, cfg.InitDefaults())
	err := cfg.Valid()
	assert.Nil(t, err)
}

func Test_AllocateTimeout(t *testing.T) {
	cfg := Config{
		NumWorkers:     10,
		DestroyTimeout: time.Second * 10,
	}
	err := cfg.Valid()

	assert.NotNil(t, err)
	assert.Equal(t, "pool.AllocateTimeout must be set", err.Error())
}

func Test_DestroyTimeout(t *testing.T) {
	cfg := Config{
		NumWorkers:      10,
		AllocateTimeout: time.Second,
	}
	err := cfg.Valid()

	assert.NotNil(t, err)
	assert.Equal(t, "pool.DestroyTimeout must be set", err.Error())
}
