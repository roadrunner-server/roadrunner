package roadrunner

import (
	"sync"
	"time"
)

const (
	// EventStderrOutput - is triggered when worker sends data into stderr. The context
	// is error message ([]byte).
	EventStderrOutput = 1900

	// WaitDuration - for how long error buffer should attempt to aggregate error messages
	// before merging output together since lastError update (required to keep error update together).
	WaitDuration = 100 * time.Millisecond
)

// thread safe errBuffer
type errBuffer struct {
	mu     sync.Mutex
	buf    []byte
	last   int
	wait   *time.Timer
	update chan interface{}
	stop   chan interface{}
	lsn    func(event int, ctx interface{})
}

func newErrBuffer() *errBuffer {
	eb := &errBuffer{
		buf:    make([]byte, 0),
		update: make(chan interface{}),
		wait:   time.NewTimer(WaitDuration),
		stop:   make(chan interface{}),
	}

	go func() {
		for {
			select {
			case <-eb.update:
				eb.wait.Reset(WaitDuration)
			case <-eb.wait.C:
				eb.mu.Lock()
				if len(eb.buf) > eb.last {
					if eb.lsn != nil {
						eb.lsn(EventStderrOutput, eb.buf[eb.last:])
						eb.buf = eb.buf[0:0]
					}

					eb.last = len(eb.buf)
				}
				eb.mu.Unlock()
			case <-eb.stop:
				eb.wait.Stop()

				eb.mu.Lock()
				if len(eb.buf) > eb.last {
					if eb.lsn != nil {
						eb.lsn(EventStderrOutput, eb.buf[eb.last:])
					}

					eb.last = len(eb.buf)
				}
				eb.mu.Unlock()
				return
			}
		}
	}()

	return eb
}

// Listen attaches error stream even listener.
func (eb *errBuffer) Listen(l func(event int, ctx interface{})) {
	eb.mu.Lock()
	eb.lsn = l
	eb.mu.Unlock()
}

// Len returns the number of buf of the unread portion of the errBuffer;
// buf.Len() == len(buf.Bytes()).
func (eb *errBuffer) Len() int {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// currently active message
	return len(eb.buf)
}

// Write appends the contents of pool to the errBuffer, growing the errBuffer as
// needed. The return value n is the length of pool; err is always nil.
func (eb *errBuffer) Write(p []byte) (int, error) {
	eb.mu.Lock()
	eb.buf = append(eb.buf, p...)
	eb.update <- nil
	eb.mu.Unlock()

	return len(p), nil
}

// Strings fetches all errBuffer data into string.
func (eb *errBuffer) String() string {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	return string(eb.buf)
}

// Close aggregation timer.
func (eb *errBuffer) Close() error {
	close(eb.stop)
	return nil
}
