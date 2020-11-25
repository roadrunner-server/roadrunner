package roadrunner

import (
	"context"
	"os/exec"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetState(t *testing.T) {
	ctx := context.Background()
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")

	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	go func() {
		assert.NoError(t, w.Wait(ctx))
		assert.Equal(t, StateStopped, w.State().Value())
	}()

	assert.NoError(t, err)
	assert.NotNil(t, w)

	assert.Equal(t, StateReady, w.State().Value())
	err = w.Stop(ctx)
	if err != nil {
		t.Errorf("error stopping the WorkerProcess: error %v", err)
	}
}

func Test_Kill(t *testing.T) {
	ctx := context.Background()
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")

	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.Error(t, w.Wait(ctx))
		// TODO changed from stopped, discuss
		assert.Equal(t, StateErrored, w.State().Value())
	}()

	assert.NoError(t, err)
	assert.NotNil(t, w)

	assert.Equal(t, StateReady, w.State().Value())
	err = w.Kill()
	if err != nil {
		t.Errorf("error killing the WorkerProcess: error %v", err)
	}
	wg.Wait()
}

func Test_OnStarted(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "broken", "pipes")
	assert.Nil(t, cmd.Start())

	w, err := InitBaseWorker(cmd)
	assert.Nil(t, w)
	assert.NotNil(t, err)

	assert.Equal(t, "can't attach to running process", err.Error())
}
