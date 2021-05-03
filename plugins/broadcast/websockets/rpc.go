package websockets

type rpcService struct {
	svc *Service
}

// Subscribe subscribes broadcast client to the given topic ahead of any websocket connections.
func (r *rpcService) Subscribe(topic string, ok *bool) error {
	*ok = true
	return r.svc.client.Subscribe(topic)
}

// SubscribePattern subscribes broadcast client to
func (r *rpcService) SubscribePattern(pattern string, ok *bool) error {
	*ok = true
	return r.svc.client.SubscribePattern(pattern)
}
