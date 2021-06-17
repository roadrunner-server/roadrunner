package broadcast

import "github.com/spiral/roadrunner/v2/pkg/interface/pubsub"

type Broadcaster interface {
	GetDriver(key string) (pubsub.SubReader, error)
}
