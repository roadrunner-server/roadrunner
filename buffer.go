package roadrunner

import (
	"bytes"
	"sync"
)

// thread safe buffer
type buffer struct {
	mu     sync.Mutex
	buffer *bytes.Buffer
}

// Len returns the number of bytes of the unread portion of the buffer;
// b.Len() == len(b.Bytes()).
func (b *buffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buffer.Len()
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. The return value n is the length of p; err is always nil. If the
// buffer becomes too large, Write will panic with ErrTooLarge.
func (b *buffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buffer.Write(p)
}

// Strings fetches all buffer data into string.
func (b *buffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buffer.String()
}
