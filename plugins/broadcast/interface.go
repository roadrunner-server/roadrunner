package broadcast

import "github.com/spiral/roadrunner/v2/common/pubsub"

type Broadcaster interface {
	GetDriver(key string) (pubsub.SubReader, error)
}
