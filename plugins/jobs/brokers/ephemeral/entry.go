package ephemeral

type entry struct {
	id       string
	priority uint64
}

func (e *entry) ID() string {
	return e.id
}

func (e *entry) Priority() uint64 {
	return e.priority
}

func (e *entry) Ask() {
	// no-op
}

func (e *entry) Nack() {
	// no-op
}

func (e *entry) Payload() []byte {
	panic("implement me")
}
