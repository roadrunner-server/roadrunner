package util

import (
	"sync/atomic"
	"time"
)

// FastTime provides current unix time using specified resolution with reduced number of syscalls.
type FastTime struct {
	last   int64
	ticker *time.Ticker
}

// NewFastTime returns new time provider with given resolution.
func NewFastTime(resolution time.Duration) *FastTime {
	ft := &FastTime{
		last:   time.Now().UnixNano(),
		ticker: time.NewTicker(resolution),
	}

	go ft.run()

	return ft
}

// Stop ticking.
func (ft *FastTime) Stop() {
	ft.ticker.Stop()
}

// UnixNano returns current timestamps. Value might be delayed after current time by specified resolution.
func (ft *FastTime) UnixNano() int64 {
	return atomic.LoadInt64(&ft.last)
}

// consume time values over given resolution.
func (ft *FastTime) run() {
	for range ft.ticker.C {
		atomic.StoreInt64(&ft.last, time.Now().UnixNano())
	}
}
