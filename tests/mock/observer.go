// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// API modeled on go.uber.org/zap/zaptest/observer, reimplemented for log/slog.

package mocklogger

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// LoggedRecord is a captured log record.
type LoggedRecord struct {
	Time    time.Time
	Level   slog.Level
	Message string
	Attrs   []slog.Attr
}

// ContextMap returns a map of all attributes attached to the record.
func (r LoggedRecord) ContextMap() map[string]any {
	m := make(map[string]any, len(r.Attrs))
	for _, a := range r.Attrs {
		m[a.Key] = a.Value.Any()
	}

	return m
}

// ObservedLogs is a concurrency-safe, ordered collection of captured logs.
type ObservedLogs struct {
	mu   sync.RWMutex
	logs []LoggedRecord
}

// Len returns the number of items in the collection.
func (o *ObservedLogs) Len() int {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return len(o.logs)
}

// All returns a copy of all the observed logs.
func (o *ObservedLogs) All() []LoggedRecord {
	o.mu.RLock()
	defer o.mu.RUnlock()

	all := make([]LoggedRecord, len(o.logs))
	copy(all, o.logs)

	return all
}

// TakeAll returns a copy of all the observed logs and truncates the observed slice.
func (o *ObservedLogs) TakeAll() []LoggedRecord {
	o.mu.Lock()
	defer o.mu.Unlock()

	ret := o.logs
	o.logs = nil

	return ret
}

// FilterMessageSnippet filters entries to those whose message contains the snippet.
func (o *ObservedLogs) FilterMessageSnippet(snippet string) *ObservedLogs {
	return o.filter(func(r LoggedRecord) bool {
		return strings.Contains(r.Message, snippet)
	})
}

// FilterMessage filters entries to those with the exact message.
func (o *ObservedLogs) FilterMessage(msg string) *ObservedLogs {
	return o.filter(func(r LoggedRecord) bool {
		return r.Message == msg
	})
}

func (o *ObservedLogs) filter(keep func(LoggedRecord) bool) *ObservedLogs {
	o.mu.RLock()
	defer o.mu.RUnlock()

	var filtered []LoggedRecord
	for _, r := range o.logs {
		if keep(r) {
			filtered = append(filtered, r)
		}
	}

	return &ObservedLogs{logs: filtered}
}

func (o *ObservedLogs) add(r LoggedRecord) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.logs = append(o.logs, r)
}

// NewObserver returns an slog.Handler that records all logs at or above the
// given level, along with the ObservedLogs collecting them.
func NewObserver(level slog.Leveler) (slog.Handler, *ObservedLogs) {
	logs := &ObservedLogs{}

	return &observerHandler{
		logs:  logs,
		level: level,
	}, logs
}

// observerHandler is an slog.Handler that appends every record to ObservedLogs.
type observerHandler struct {
	logs  *ObservedLogs
	level slog.Leveler
	attrs []slog.Attr
	group string
}

func (h *observerHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *observerHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := make([]slog.Attr, 0, len(h.attrs)+r.NumAttrs())
	attrs = append(attrs, h.attrs...)

	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, h.qualify(a))
		return true
	})

	h.logs.add(LoggedRecord{
		Time:    r.Time,
		Level:   r.Level,
		Message: r.Message,
		Attrs:   attrs,
	})

	return nil
}

func (h *observerHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nh := *h
	nh.attrs = make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	nh.attrs = append(nh.attrs, h.attrs...)

	for _, a := range attrs {
		nh.attrs = append(nh.attrs, h.qualify(a))
	}

	return &nh
}

func (h *observerHandler) WithGroup(name string) slog.Handler {
	nh := *h
	if h.group != "" {
		nh.group = h.group + "." + name
	} else {
		nh.group = name
	}

	return &nh
}

// qualify prefixes the attribute key with the active group path.
func (h *observerHandler) qualify(a slog.Attr) slog.Attr {
	if h.group == "" {
		return a
	}

	return slog.Attr{Key: h.group + "." + a.Key, Value: a.Value}
}
