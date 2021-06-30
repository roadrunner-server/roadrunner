package ephemeral

type queue struct{}

func newQueue() *queue {
	return &queue{}
}
