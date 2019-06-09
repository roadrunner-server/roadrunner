package http

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func createStatsListener() statsListener {
	return statsListener{
		stats: &ServiceStats{},
	};
}

func TestHandler_Stats_Listener_EventResponse(t *testing.T) {
	listener := createStatsListener();

	listener.listener(EventResponse, &ResponseEvent{Request: nil, Response: nil, start: time.Now(), elapsed: 0})
	listener.listener(EventResponse, &ResponseEvent{Request: nil, Response: nil, start: time.Now(), elapsed: 0})

	assert.Equal(t, uint64(2), listener.stats.Accepted)
	assert.Equal(t, uint64(2), listener.stats.Success)
	assert.Equal(t, uint64(0), listener.stats.Error)
}

func TestHandler_Stats_Listener_EventError(t *testing.T) {
	listener := createStatsListener();

	listener.listener(EventError, &ErrorEvent{Request: nil, Error: nil, start: time.Now(), elapsed: 0})
	listener.listener(EventError, &ErrorEvent{Request: nil, Error: nil, start: time.Now(), elapsed: 0})

	assert.Equal(t, uint64(2), listener.stats.Accepted)
	assert.Equal(t, uint64(0), listener.stats.Success)
	assert.Equal(t, uint64(2), listener.stats.Error)
}
