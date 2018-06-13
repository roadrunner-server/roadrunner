package roadrunner

import (
	"bytes"
	"sync"
)

// thread safe errBuffer
type errBuffer struct {
	mu     sync.Mutex
	buffer *bytes.Buffer
}

// Len returns the number of bytes of the unread portion of the errBuffer;
// b.Len() == len(b.Bytes()).
func (b *errBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buffer.Len()
}

// Write appends the contents of p to the errBuffer, growing the errBuffer as
// needed. The return value n is the length of p; err is always nil. If the
// errBuffer becomes too large, Write will panic with ErrTooLarge.
func (b *errBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buffer.Write(p)
}

// Strings fetches all errBuffer data into string.
func (b *errBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buffer.String()
}
