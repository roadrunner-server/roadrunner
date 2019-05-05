package service

import (
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

type testService struct {
	mu           sync.Mutex
	waitForServe chan interface{}
	delay        time.Duration
	ok           bool
	cfg          Config
	c            Container
	cfgE, serveE error
	done         chan interface{}
}

func (t *testService) Init(cfg Config, c Container) (enabled bool, err error) {
	t.cfg = cfg
	t.c = c
	t.done = make(chan interface{})
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

	<-t.done
	return nil
}

func (t *testService) Stop() {
	close(t.done)
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
	vars := make(map[string]interface{})
	json.Unmarshal([]byte(cfg.cfg), &vars)

	v, ok := vars[name]
	if !ok {
		return nil
	}

	d, _ := json.Marshal(v)
	return &testCfg{cfg: string(d)}
}
func (cfg *testCfg) Unmarshal(out interface{}) error { return json.Unmarshal([]byte(cfg.cfg), out) }

// Config defines RPC service config.
type dConfig struct {
	// Indicates if RPC connection is enabled.
	Value string
}

// Hydrate must populate Config values using given Config source. Must return error if Config is not valid.
func (c *dConfig) Hydrate(cfg Config) error {
	return cfg.Unmarshal(c)
}

// InitDefaults allows to init blank config with pre-defined set of default values.
func (c *dConfig) InitDefaults() error {
	c.Value = "default"

	return nil
}

type dService struct {
	Cfg *dConfig
}

func (s *dService) Init(cfg *dConfig) (bool, error) {
	s.Cfg = cfg
	return true, nil
}

func TestContainer_Register(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := NewContainer(logger)
	c.Register("test", &testService{})

	assert.Equal(t, 0, len(hook.Entries))
}

func TestContainer_Has(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := NewContainer(logger)
	c.Register("test", &testService{})

	assert.Equal(t, 0, len(hook.Entries))

	assert.True(t, c.Has("test"))
	assert.False(t, c.Has("another"))
}

func TestContainer_Get(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := NewContainer(logger)
	c.Register("test", &testService{})
	assert.Equal(t, 0, len(hook.Entries))

	s, st := c.Get("test")
	assert.IsType(t, &testService{}, s)
	assert.Equal(t, StatusInactive, st)

	s, st = c.Get("another")
	assert.Nil(t, s)
	assert.Equal(t, StatusUndefined, st)
}

func TestContainer_Stop_NotStarted(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := NewContainer(logger)
	c.Register("test", &testService{})
	assert.Equal(t, 0, len(hook.Entries))

	c.Stop()
}

func TestContainer_Configure(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	svc := &testService{ok: true}

	c := NewContainer(logger)
	c.Register("test", svc)
	assert.Equal(t, 0, len(hook.Entries))

	assert.NoError(t, c.Init(&testCfg{`{"test":"something"}`}))

	s, st := c.Get("test")
	assert.IsType(t, &testService{}, s)
	assert.Equal(t, StatusOK, st)
}

func TestContainer_Init_Default(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	svc := &dService{}

	c := NewContainer(logger)
	c.Register("test", svc)
	assert.Equal(t, 0, len(hook.Entries))

	assert.NoError(t, c.Init(&testCfg{`{}`}))

	s, st := c.Get("test")
	assert.IsType(t, &dService{}, s)
	assert.Equal(t, StatusOK, st)

	assert.Equal(t, "default", svc.Cfg.Value)
}

func TestContainer_Init_Default_Overwrite(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	svc := &dService{}

	c := NewContainer(logger)
	c.Register("test", svc)
	assert.Equal(t, 0, len(hook.Entries))

	assert.NoError(t, c.Init(&testCfg{`{"test":{"value": "something"}}`}))

	s, st := c.Get("test")
	assert.IsType(t, &dService{}, s)
	assert.Equal(t, StatusOK, st)

	assert.Equal(t, "something", svc.Cfg.Value)
}

func TestContainer_ConfigureNull(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	svc := &testService{ok: true}

	c := NewContainer(logger)
	c.Register("test", svc)
	assert.Equal(t, 0, len(hook.Entries))

	assert.NoError(t, c.Init(&testCfg{`{"another":"something"}`}))
	assert.Equal(t, 1, len(hook.Entries))

	s, st := c.Get("test")
	assert.IsType(t, &testService{}, s)
	assert.Equal(t, StatusInactive, st)
}

func TestContainer_ConfigureDisabled(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	svc := &testService{ok: false}

	c := NewContainer(logger)
	c.Register("test", svc)
	assert.Equal(t, 0, len(hook.Entries))

	assert.NoError(t, c.Init(&testCfg{`{"test":"something"}`}))
	assert.Equal(t, 1, len(hook.Entries))

	s, st := c.Get("test")
	assert.IsType(t, &testService{}, s)
	assert.Equal(t, StatusInactive, st)
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
	assert.Equal(t, 0, len(hook.Entries))

	err := c.Init(&testCfg{`{"test":"something"}`})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configure error")
	assert.Contains(t, err.Error(), "test")

	s, st := c.Get("test")
	assert.IsType(t, &testService{}, s)
	assert.Equal(t, StatusInactive, st)
}

func TestContainer_ConfigureTwice(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	svc := &testService{ok: true}

	c := NewContainer(logger)
	c.Register("test", svc)
	assert.Equal(t, 0, len(hook.Entries))

	assert.NoError(t, c.Init(&testCfg{`{"test":"something"}`}))
	assert.Error(t, c.Init(&testCfg{`{"test":"something"}`}))
}

func TestContainer_ServeEmptyContainer(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	svc := &testService{ok: true}

	c := NewContainer(logger)
	c.Register("test", svc)
	assert.Equal(t, 0, len(hook.Entries))

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
	assert.Equal(t, 0, len(hook.Entries))
	assert.NoError(t, c.Init(&testCfg{`{"test":"something"}`}))

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
	assert.Equal(t, 0, len(hook.Entries))
	assert.NoError(t, c.Init(&testCfg{`{"test":"something"}`}))

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
	assert.Equal(t, 0, len(hook.Entries))
	assert.NoError(t, c.Init(&testCfg{`{"test":"something", "test2":"something-else"}`}))

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
}

type testInitA struct{}

func (t *testInitA) Init() error {
	return nil
}

type testInitB struct{}

func (t *testInitB) Init() (int, error) {
	return 0, nil
}

func TestContainer_InitErrorA(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := NewContainer(logger)
	c.Register("test", &testInitA{})

	assert.Error(t, c.Init(&testCfg{`{"test":"something", "test2":"something-else"}`}))
}

func TestContainer_InitErrorB(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := NewContainer(logger)
	c.Register("test", &testInitB{})

	assert.Error(t, c.Init(&testCfg{`{"test":"something", "test2":"something-else"}`}))
}

type testInitC struct{}

func TestContainer_NoInit(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := NewContainer(logger)
	c.Register("test", &testInitC{})

	assert.NoError(t, c.Init(&testCfg{`{"test":"something", "test2":"something-else"}`}))
}

type testInitD struct {
	c *testInitC
}

type DCfg struct {
	V string
}

// Hydrate must populate Config values using given Config source. Must return error if Config is not valid.
func (c *DCfg) Hydrate(cfg Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	if c.V == "fail" {
		return errors.New("failed config")
	}

	return nil
}

func (t *testInitD) Init(r *testInitC, c Container, cfg *DCfg) (bool, error) {
	if r == nil {
		return false, errors.New("unable to find testInitC")
	}

	if c == nil {
		return false, errors.New("unable to find Container")
	}

	if cfg.V != "ok" {
		return false, errors.New("invalid config")
	}

	return false, nil
}

func TestContainer_InitDependency(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := NewContainer(logger)
	c.Register("test", &testInitC{})
	c.Register("test2", &testInitD{})

	assert.NoError(t, c.Init(&testCfg{`{"test":"something", "test2":{"v":"ok"}}`}))
}

func TestContainer_InitDependencyFail(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := NewContainer(logger)
	c.Register("test", &testInitC{})
	c.Register("test2", &testInitD{})

	assert.Error(t, c.Init(&testCfg{`{"test":"something", "test2":{"v":"fail"}}`}))
}

func TestContainer_InitDependencyEmpty(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := NewContainer(logger)
	c.Register("test2", &testInitD{})

	assert.Contains(t, c.Init(&testCfg{`{"test2":{"v":"ok"}}`}).Error(), "testInitC")
}
