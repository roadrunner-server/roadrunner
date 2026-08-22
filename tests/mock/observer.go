package mocklogger

import (
	"context"
	"log/slog"
	"strings"
	"sync"
)

// LoggedEntry is a representation of a log record captured by the observer.
type LoggedEntry struct {
	Level   slog.Level
	Message string
}

// ObservedLogs is a concurrency-safe, ordered collection of observed logs.
type ObservedLogs struct {
	mu   sync.RWMutex
	logs []LoggedEntry
}

// Len returns the number of items in the collection.
func (o *ObservedLogs) Len() int {
	o.mu.RLock()
	n := len(o.logs)
	o.mu.RUnlock()
	return n
}

// FilterMessageSnippet returns the entries whose message contains the snippet.
func (o *ObservedLogs) FilterMessageSnippet(snippet string) *ObservedLogs {
	o.mu.RLock()
	defer o.mu.RUnlock()

	var filtered []LoggedEntry
	for _, entry := range o.logs {
		if strings.Contains(entry.Message, snippet) {
			filtered = append(filtered, entry)
		}
	}

	return &ObservedLogs{logs: filtered}
}

func (o *ObservedLogs) add(entry LoggedEntry) {
	o.mu.Lock()
	o.logs = append(o.logs, entry)
	o.mu.Unlock()
}

// observerHandler is an slog.Handler that captures the level and the message of
// every record. Attributes are not recorded, nothing asserts on them.
type observerHandler struct {
	level slog.Level
	logs  *ObservedLogs
}

// NewObserverHandler creates a new slog.Handler that buffers logs in memory.
func NewObserverHandler(level slog.Level) (slog.Handler, *ObservedLogs) {
	ol := &ObservedLogs{}
	return &observerHandler{
		level: level,
		logs:  ol,
	}, ol
}

func (h *observerHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *observerHandler) Handle(_ context.Context, r slog.Record) error {
	h.logs.add(LoggedEntry{
		Level:   r.Level,
		Message: r.Message,
	})
	return nil
}

func (h *observerHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *observerHandler) WithGroup(_ string) slog.Handler {
	return h
}
