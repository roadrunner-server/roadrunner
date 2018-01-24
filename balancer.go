package roadrunner

//
//import (
//	"os/exec"
//	"sync"
//)
//
//// Swapper provides ability to perform hot-swap between 2 worker pools.
//type Swapper struct {
//	mu   sync.Mutex // protects pool hot swapping
//	pool *Pool      // pool to work for user commands
//}
//
//// Swap initiates underlying pool of workers and replaces old one.
//func (b *Swapper) Swap(cmd func() *exec.Cmd, factory Factory, cfg Config) error {
//	var (
//		err  error
//		prev *Pool
//		pool *Pool
//	)
//
//	prev = b.pool
//	if pool, err = NewPool(cmd, factory, cfg); err != nil {
//		return err
//	}
//
//	if prev != nil {
//		go func() {
//			prev.Close()
//		}()
//	}
//
//	b.mu.Lock()
//	b.pool = pool
//	b.mu.Unlock()
//
//	return nil
//}
//
//// Exec one task with given payload and context, returns result and context
//// or error. Must not be used once pool is being destroyed.
//func (b *Swapper) Exec(payload []byte, ctx interface{}) (resp []byte, rCtx []byte, err error) {
//	b.mu.Lock()
//	pool := b.pool
//	b.mu.Unlock()
//
//	if pool == nil {
//		panic("what")
//	}
//
//	return pool.Exec(payload, ctx)
//}
//
//// Workers return list of active workers.
//func (b *Swapper) Workers() []*Worker {
//	b.mu.Lock()
//	pool := b.pool
//	b.mu.Unlock()
//
//	return pool.Workers()
//}
//
//// Close closes underlying pool.
//func (b *Swapper) Close() {
//	b.mu.Lock()
//	defer b.mu.Unlock()
//
//	b.pool.Close()
//	b.pool = nil
//}
