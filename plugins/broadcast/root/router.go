package broadcast

//import "github.com/gobwas/glob"

// Router performs internal message routing to multiple subscribers.
type Router struct {
	wildcard map[string]wildcard
	routes   map[string][]chan *Message
}

// wildcard handles number of topics via glob pattern.
type wildcard struct {
	//glob     glob.Glob
	upstream []chan *Message
}

// helper for blocking join/leave flow
type subscriber struct {
	upstream chan *Message
	done     chan error
	topics   []string
	pattern  string
}

// NewRouter creates new topic and pattern router.
func NewRouter() *Router {
	return &Router{
		wildcard: make(map[string]wildcard),
		routes:   make(map[string][]chan *Message),
	}
}

// Dispatch to all connected topics.
func (r *Router) Dispatch(msg *Message) {
	for _, w := range r.wildcard {
		if w.glob.Match(msg.Topic) {
			for _, upstream := range w.upstream {
				upstream <- msg
			}
		}
	}

	if routes, ok := r.routes[msg.Topic]; ok {
		for _, upstream := range routes {
			upstream <- msg
		}
	}
}

// Subscribe to topic and return list of newly assigned topics.
func (r *Router) Subscribe(upstream chan *Message, topics ...string) (newTopics []string) {
	newTopics = make([]string, 0)
	for _, topic := range topics {
		if _, ok := r.routes[topic]; !ok {
			r.routes[topic] = []chan *Message{upstream}
			if !r.collapsed(topic) {
				newTopics = append(newTopics, topic)
			}
			continue
		}

		joined := false
		for _, up := range r.routes[topic] {
			if up == upstream {
				joined = true
				break
			}
		}

		if !joined {
			r.routes[topic] = append(r.routes[topic], upstream)
		}
	}

	return newTopics
}

// Unsubscribe from given list of topics and return list of topics which are no longer claimed.
func (r *Router) Unsubscribe(upstream chan *Message, topics ...string) (dropTopics []string) {
	dropTopics = make([]string, 0)
	for _, topic := range topics {
		if _, ok := r.routes[topic]; !ok {
			// no such topic, ignore
			continue
		}

		for i := range r.routes[topic] {
			if r.routes[topic][i] == upstream {
				r.routes[topic] = append(r.routes[topic][:i], r.routes[topic][i+1:]...)
				break
			}
		}

		if len(r.routes[topic]) == 0 {
			delete(r.routes, topic)

			// standalone empty subscription
			if !r.collapsed(topic) {
				dropTopics = append(dropTopics, topic)
			}
		}
	}

	return dropTopics
}

// SubscribePattern subscribes to glob parent and return true and return array of newly added patterns. Error in
// case if blob is invalid.
func (r *Router) SubscribePattern(upstream chan *Message, pattern string) (newPatterns []string, err error) {
	if w, ok := r.wildcard[pattern]; ok {
		joined := false
		for _, up := range w.upstream {
			if up == upstream {
				joined = true
				break
			}
		}

		if !joined {
			w.upstream = append(w.upstream, upstream)
		}

		return nil, nil
	}

	g, err := glob.Compile(pattern)
	if err != nil {
		return nil, err
	}

	r.wildcard[pattern] = wildcard{glob: g, upstream: []chan *Message{upstream}}

	return []string{pattern}, nil
}

// UnsubscribePattern unsubscribe from the pattern and returns an array of patterns which are no longer claimed.
func (r *Router) UnsubscribePattern(upstream chan *Message, pattern string) (dropPatterns []string) {
	// todo: store and return collapsed topics

	w, ok := r.wildcard[pattern]
	if !ok {
		// no such pattern
		return nil
	}

	for i, up := range w.upstream {
		if up == upstream {
			w.upstream[i] = w.upstream[len(w.upstream)-1]
			w.upstream[len(w.upstream)-1] = nil
			w.upstream = w.upstream[:len(w.upstream)-1]

			if len(w.upstream) == 0 {
				delete(r.wildcard, pattern)
				return []string{pattern}
			}
		}
	}

	return nil
}

func (r *Router) collapsed(topic string) bool {
	for _, w := range r.wildcard {
		if w.glob.Match(topic) {
			return true
		}
	}

	return false
}
