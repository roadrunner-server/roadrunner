package roadrunner

import (
	"bufio"
)

// Payload carries binary header and body to workers and
// back to the server.
type Payload struct {
	// Context represent payload context, might be omitted.
	Context []byte

	// body contains binary payload to be processed by worker.
	Body []byte

	// attached when worker responds with the stream
	stream *bufio.Reader

	// close callback will be called when payload is closed
	cc func()
}

// String returns payload body as string
func (p *Payload) String() string {
	return string(p.Body)
}

// Stream returns true is payload is streaming.
func (p *Payload) Stream() bool {
	return p.stream != nil
}

// Stream returns associated stream.
func (p *Payload) Read(d []byte) (n int, err error) {
	return p.stream.Read(d)
}

// Close closes underlying stream and notifies stream end watchers.
func (p *Payload) Close() {
	if p.cc != nil {
		p.cc()
	}
}
