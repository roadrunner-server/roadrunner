package roadrunner

// Payload carries binary header and body to workers and
// back to the server.
type Payload struct {
	// Head represent payload context, might be omitted
	Head []byte

	// Body contains binary payload to be processed by worker
	Body []byte
}

// String returns payload body as string
func (p *Payload) String() string {
	return string(p.Body)
}
