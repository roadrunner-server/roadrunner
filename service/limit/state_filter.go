package limit

import (
	"github.com/spiral/roadrunner"
	"time"
)

type stateFilter struct {
	prev map[*roadrunner.Worker]state
	next map[*roadrunner.Worker]state
}

type state struct {
	state    int64
	numExecs int64
	since    time.Time
}

func newStateFilter() *stateFilter {
	return &stateFilter{
		prev: make(map[*roadrunner.Worker]state),
		next: make(map[*roadrunner.Worker]state),
	}
}

// add new worker to be watched
func (sw *stateFilter) push(w *roadrunner.Worker) {
	sw.next[w] = state{state: w.State().Value(), numExecs: w.State().NumExecs()}
}

// update worker states.
func (sw *stateFilter) sync(t time.Time) {
	for w := range sw.prev {
		if _, ok := sw.next[w]; !ok {
			delete(sw.prev, w)
		}
	}

	for w, s := range sw.next {
		ps, ok := sw.prev[w]
		if !ok || ps.state != s.state || ps.numExecs != s.numExecs {
			sw.prev[w] = state{state: s.state, numExecs: s.numExecs, since: t}
		}

		delete(sw.next, w)
	}
}

// find all workers which spend given amount of time in a specific state.
func (sw *stateFilter) find(state int64, since time.Time) (workers []*roadrunner.Worker) {
	for w, s := range sw.prev {
		if s.state == state && s.since.Before(since) {
			workers = append(workers, w)
		}
	}

	return
}
