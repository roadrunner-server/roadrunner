package roadrunner

import (
	"bytes"
	"sync"
)

// EventStderrOutput - is triggered when worker sends data into stderr. The context is error message ([]byte).
const EventStderrOutput = 1900

// thread safe errBuffer
type errBuffer struct {
	mu  sync.Mutex
	buf []byte
	off int
	lsn func(event int, ctx interface{})
}

func newErrBuffer() *errBuffer {
	buf := &errBuffer{buf: make([]byte, 0)}
	return buf
}

// Listen attaches error stream even listener.
func (eb *errBuffer) Listen(l func(event int, ctx interface{})) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.lsn = l
}

// Len returns the number of buf of the unread portion of the errBuffer;
// buf.Len() == len(buf.Bytes()).
func (eb *errBuffer) Len() int {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// currently active message
	return len(eb.buf) - eb.off
}

// Write appends the contents of p to the errBuffer, growing the errBuffer as
// needed. The return value n is the length of p; err is always nil.
func (eb *errBuffer) Write(p []byte) (int, error) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.buf = append(eb.buf, p...)
	for msg := eb.fetchMsg(); msg != nil; msg = eb.fetchMsg() {
		eb.lsn(EventStderrOutput, msg)
	}

	return len(p), nil
}

// Strings fetches all errBuffer data into string.
func (eb *errBuffer) String() string {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	return string(eb.buf[eb.off:])
}

func (eb *errBuffer) fetchMsg() []byte {
	if i := bytes.Index(eb.buf[eb.off:], []byte{10, 10}); i != -1 {
		eb.off += i + 2
		return eb.buf[eb.off-i-2 : eb.off]
	}

	return nil
}
