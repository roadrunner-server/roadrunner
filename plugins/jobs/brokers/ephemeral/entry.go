package ephemeral

type entry struct {
	id string
}

func (e *entry) ID() string {
	return e.id
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
