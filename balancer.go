package roadrunner

import (
	"os/exec"
	"sync"
)

// Balancer provides ability to perform hot-swap between 2 worker pools.
type Balancer struct {
	mu   sync.Mutex // protects pool hot swapping
	pool *Pool      // pool to work for user commands
}

// Spawn initiates underlying pool of workers and replaced old one.
func (b *Balancer) Spawn(cmd func() *exec.Cmd, factory Factory, cfg Config) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	var (
		err error
		old *Pool
	)

	old = b.pool
	if b.pool, err = NewPool(cmd, factory, cfg); err != nil {
		return err
	}

	if old != nil {
		go func() {
			old.Close()
		}()
	}

	return nil
}

// Exec one task with given payload and context, returns result and context
// or error. Must not be used once pool is being destroyed.
func (b *Balancer) Exec(payload []byte, ctx interface{}) (resp []byte, rCtx []byte, err error) {
	b.mu.Lock()
	pool := b.pool
	b.mu.Unlock()

	return pool.Exec(payload, ctx)
}

// Workers return list of active workers.
func (b *Balancer) Workers() []*Worker {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.pool.Workers()
}

// Close closes underlying pool.
func (b *Balancer) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.pool.Close()
	b.pool = nil
}
