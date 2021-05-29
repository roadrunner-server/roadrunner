package channel

// Hub used as a channel between two or more plugins
// this is not a PUBLIC plugin, API might be changed at any moment
type Hub interface {
	FromWorker() chan interface{}
	ToWorker() chan interface{}
}
