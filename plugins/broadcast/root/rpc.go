package broadcast

import "golang.org/x/sync/errgroup"

type rpcService struct {
	svc *Service
}

// Publish Messages.
func (r *rpcService) Publish(msg []*Message, ok *bool) error {
	*ok = true
	return r.svc.Publish(msg...)
}

// Publish Messages in async mode. Blocks until get an err or nil from publish
func (r *rpcService) PublishAsync(msg []*Message, ok *bool) error {
	*ok = true
	g := &errgroup.Group{}

	g.Go(func() error {
		return r.svc.Publish(msg...)
	})

	return g.Wait()
}
