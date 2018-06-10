package service

import (
	"testing"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/sirupsen/logrus"
	"encoding/json"
	"errors"
	"time"
	"sync"
)

type testService struct {
	mu           sync.Mutex
	waitForServe chan interface{}
	delay        time.Duration
	ok           bool
	cfg          Config
	c            Container
	cfgE, serveE error
	serving      chan interface{}
}

func (t *testService) Configure(cfg Config, c Container) (enabled bool, err error) {
	t.cfg = cfg
	t.c = c
	return t.ok, t.cfgE
}

func (t *testService) Serve() error {
	time.Sleep(t.delay)

	if t.serveE != nil {
		return t.serveE
	}

	if c := t.waitChan(); c != nil {
		close(c)
		t.setChan(nil)
	}

	t.mu.Lock()
	t.serving = make(chan interface{})
	t.mu.Unlock()

	<-t.serving

	return nil
}

func (t *testService) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	close(t.serving)
}

func (t *testService) waitChan() chan interface{} {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.waitForServe
}

func (t *testService) setChan(c chan interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.waitForServe = c
}

type testCfg struct{ cfg string }

func (cfg *testCfg) Get(name string) Config {
	vars := make(map[string]string)
	json.Unmarshal([]byte(cfg.cfg), &vars)

	v, ok := vars[name]
	if !ok {
		return nil
	}

	return &testCfg{cfg: v}
}
func (cfg *testCfg) Unmarshal(out interface{}) error { return json.Unmarshal([]byte(cfg.cfg), out) }

func TestContainer_Register(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := NewContainer(logger)
	c.Register("test", &testService{})

	assert.Equal(t, 1, len(hook.Entries))
}

func TestContainer_Has(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := NewContainer(logger)
	c.Register("test", &testService{})

	assert.Equal(t, 1, len(hook.Entries))

	assert.True(t, c.Has("test"))
	assert.False(t, c.Has("another"))
}

func TestContainer_Get(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := NewContainer(logger)
	c.Register("test", &testService{})

	assert.Equal(t, 1, len(hook.Entries))

	s, st := c.Get("test")
	assert.IsType(t, &testService{}, s)
	assert.Equal(t, StatusRegistered, st)

	s, st = c.Get("another")
	assert.Nil(t, s)
	assert.Equal(t, StatusUndefined, st)
}

func TestContainer_Stop_NotStarted(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := NewContainer(logger)
	c.Register("test", &testService{})
	assert.Equal(t, 1, len(hook.Entries))

	c.Stop()
}

func TestContainer_Configure(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	svc := &testService{ok: true}

	c := NewContainer(logger)
	c.Register("test", svc)
	assert.Equal(t, 1, len(hook.Entries))

	assert.NoError(t, c.Configure(&testCfg{`{"test":"something"}`}))

	s, st := c.Get("test")
	assert.IsType(t, &testService{}, s)
	assert.Equal(t, StatusConfigured, st)
}

func TestContainer_ConfigureNull(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	svc := &testService{ok: true}

	c := NewContainer(logger)
	c.Register("test", svc)
	assert.Equal(t, 1, len(hook.Entries))

	assert.NoError(t, c.Configure(&testCfg{`{"another":"something"}`}))
	assert.Equal(t, 2, len(hook.Entries))

	s, st := c.Get("test")
	assert.IsType(t, &testService{}, s)
	assert.Equal(t, StatusRegistered, st)
}

func TestContainer_ConfigureDisabled(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	svc := &testService{ok: false}

	c := NewContainer(logger)
	c.Register("test", svc)
	assert.Equal(t, 1, len(hook.Entries))

	assert.NoError(t, c.Configure(&testCfg{`{"test":"something"}`}))
	assert.Equal(t, 1, len(hook.Entries))

	s, st := c.Get("test")
	assert.IsType(t, &testService{}, s)
	assert.Equal(t, StatusRegistered, st)
}

func TestContainer_ConfigureError(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	svc := &testService{
		ok:   false,
		cfgE: errors.New("configure error"),
	}

	c := NewContainer(logger)
	c.Register("test", svc)
	assert.Equal(t, 1, len(hook.Entries))

	err := c.Configure(&testCfg{`{"test":"something"}`})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configure error")
	assert.Contains(t, err.Error(), "test.service")

	s, st := c.Get("test")
	assert.IsType(t, &testService{}, s)
	assert.Equal(t, StatusRegistered, st)
}

func TestContainer_ConfigureTwice(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	svc := &testService{ok: true}

	c := NewContainer(logger)
	c.Register("test", svc)
	assert.Equal(t, 1, len(hook.Entries))

	assert.NoError(t, c.Configure(&testCfg{`{"test":"something"}`}))
	assert.Error(t, c.Configure(&testCfg{`{"test":"something"}`}))
}

func TestContainer_ServeEmptyContainer(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	svc := &testService{ok: true}

	c := NewContainer(logger)
	c.Register("test", svc)
	assert.Equal(t, 1, len(hook.Entries))

	assert.NoError(t, c.Serve())
	c.Stop()
}

func TestContainer_Serve(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	svc := &testService{
		ok:           true,
		waitForServe: make(chan interface{}),
	}

	c := NewContainer(logger)
	c.Register("test", svc)
	assert.Equal(t, 1, len(hook.Entries))
	assert.NoError(t, c.Configure(&testCfg{`{"test":"something"}`}))

	go func() {
		assert.NoError(t, c.Serve())
	}()

	<-svc.waitChan()

	s, st := c.Get("test")
	assert.IsType(t, &testService{}, s)
	assert.Equal(t, StatusServing, st)

	c.Stop()

	s, st = c.Get("test")
	assert.IsType(t, &testService{}, s)
	assert.Equal(t, StatusStopped, st)
}

func TestContainer_ServeError(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	svc := &testService{
		ok:           true,
		waitForServe: make(chan interface{}),
		serveE:       errors.New("serve error"),
	}

	c := NewContainer(logger)
	c.Register("test", svc)
	assert.Equal(t, 1, len(hook.Entries))
	assert.NoError(t, c.Configure(&testCfg{`{"test":"something"}`}))

	err := c.Serve()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serve error")
	assert.Contains(t, err.Error(), "test")

	s, st := c.Get("test")
	assert.IsType(t, &testService{}, s)
	assert.Equal(t, StatusStopped, st)
}

func TestContainer_ServeErrorMultiple(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	svc := &testService{
		ok:           true,
		delay:        time.Millisecond * 10,
		waitForServe: make(chan interface{}),
		serveE:       errors.New("serve error"),
	}

	svc2 := &testService{
		ok:           true,
		waitForServe: make(chan interface{}),
	}

	c := NewContainer(logger)
	c.Register("test2", svc2)
	c.Register("test", svc)
	assert.Equal(t, 2, len(hook.Entries))
	assert.NoError(t, c.Configure(&testCfg{`{"test":"something", "test2":"something-else"}`}))

	err := c.Serve()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serve error")
	assert.Contains(t, err.Error(), "test")

	s, st := c.Get("test")
	assert.IsType(t, &testService{}, s)
	assert.Equal(t, StatusStopped, st)

	s, st = c.Get("test2")
	assert.IsType(t, &testService{}, s)
	assert.Equal(t, StatusStopped, st)

	assert.Equal(t, 6, len(hook.Entries))
}
