package limit

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	rrhttp "github.com/spiral/roadrunner/service/http"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

type testCfg struct {
	httpCfg  string
	limitCfg string
	target   string
}

func (cfg *testCfg) Get(name string) service.Config {
	if name == rrhttp.ID {
		if cfg.httpCfg == "" {
			return nil
		}

		return &testCfg{target: cfg.httpCfg}
	}

	if name == ID {
		return &testCfg{target: cfg.limitCfg}
	}

	return nil
}

func (cfg *testCfg) Unmarshal(out interface{}) error {
	err := json.Unmarshal([]byte(cfg.target), out)

	if cl, ok := out.(*Config); ok {
		// to speed up tests
		cl.Interval = time.Millisecond
	}

	return err
}

func Test_Service_PidEcho(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		httpCfg: `{
			"address": ":7029",
			"workers":{
				"command": "php ../../tests/http/client.php pid pipes",
				"pool": {"numWorkers": 1}
			}
	}`,
		limitCfg: `{
			"services": {
				"http": {
					"ttl": 1
				}
			}
		}`,
	}))

	s, _ := c.Get(rrhttp.ID)
	assert.NotNil(t, s)

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("error during the Serve: error %v", err)
		}
	}()

	time.Sleep(time.Millisecond * 100)
	req, err := http.NewRequest("GET", "http://localhost:7029", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)


	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, getPID(s), string(b))

	err2 := r.Body.Close()
	if err2 != nil {
		t.Errorf("error during the body closing: error %v", err2)
	}
	c.Stop()
}

func Test_Service_ListenerPlusTTL(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		httpCfg: `{
			"address": ":7030",
			"workers":{
				"command": "php ../../tests/http/client.php pid pipes",
				"pool": {"numWorkers": 1}
			}
	}`,
		limitCfg: `{
			"services": {
				"http": {
					"ttl": 1
				}
			}
		}`,
	}))

	s, _ := c.Get(rrhttp.ID)
	assert.NotNil(t, s)

	l, _ := c.Get(ID)
	captured := make(chan interface{})
	l.(*Service).AddListener(func(event int, ctx interface{}) {
		if event == EventTTL {
			close(captured)
		}
	})

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("error during the Serve: error %v", err)
		}
	}()


	time.Sleep(time.Millisecond * 100)

	lastPID := getPID(s)

	req, err := http.NewRequest("GET", "http://localhost:7030", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)


	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, lastPID, string(b))

	<-captured

	// clean state
	req, err = http.NewRequest("GET", "http://localhost:7030?new", nil)
	assert.NoError(t, err)

	_, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)

	assert.NotEqual(t, lastPID, getPID(s))

	c.Stop()

	err2 := r.Body.Close()
	if err2 != nil {
		t.Errorf("error during the body closing: error %v", err2)
	}
}

func Test_Service_ListenerPlusIdleTTL(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		httpCfg: `{
			"address": ":7031",
			"workers":{
				"command": "php ../../tests/http/client.php pid pipes",
				"pool": {"numWorkers": 1}
			}
	}`,
		limitCfg: `{
			"services": {
				"http": {
					"idleTtl": 1
				}
			}
		}`,
	}))

	s, _ := c.Get(rrhttp.ID)
	assert.NotNil(t, s)

	l, _ := c.Get(ID)
	captured := make(chan interface{})
	l.(*Service).AddListener(func(event int, ctx interface{}) {
		if event == EventIdleTTL {
			close(captured)
		}
	})

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("error during the Serve: error %v", err)
		}
	}()


	time.Sleep(time.Millisecond * 100)

	lastPID := getPID(s)

	req, err := http.NewRequest("GET", "http://localhost:7031", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, lastPID, string(b))

	<-captured

	// clean state
	req, err = http.NewRequest("GET", "http://localhost:7031?new", nil)
	assert.NoError(t, err)

	_, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)

	assert.NotEqual(t, lastPID, getPID(s))

	c.Stop()
	err2 := r.Body.Close()
	if err2 != nil {
		t.Errorf("error during the body closing: error %v", err2)
	}
}

func Test_Service_Listener_MaxExecTTL(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		httpCfg: `{
			"address": ":7032",
			"workers":{
				"command": "php ../../tests/http/client.php stuck pipes",
				"pool": {"numWorkers": 1}
			}
	}`,
		limitCfg: `{
			"services": {
				"http": {
					"execTTL": 1
				}
			}
		}`,
	}))

	s, _ := c.Get(rrhttp.ID)
	assert.NotNil(t, s)

	l, _ := c.Get(ID)
	captured := make(chan interface{})
	l.(*Service).AddListener(func(event int, ctx interface{}) {
		if event == EventExecTTL {
			close(captured)
		}
	})

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("error during the Serve: error %v", err)
		}
	}()

	time.Sleep(time.Millisecond * 100)

	req, err := http.NewRequest("GET", "http://localhost:7032", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 500, r.StatusCode)

	<-captured

	c.Stop()
}

func Test_Service_Listener_MaxMemoryUsage(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		httpCfg: `{
			"address": ":7033",
			"workers":{
				"command": "php ../../tests/http/client.php memleak pipes",
				"pool": {"numWorkers": 1}
			}
	}`,
		limitCfg: `{
			"services": {
				"http": {
					"maxMemory": 10
				}
			}
		}`,
	}))

	s, _ := c.Get(rrhttp.ID)
	assert.NotNil(t, s)

	l, _ := c.Get(ID)
	captured := make(chan interface{})
	once := false
	l.(*Service).AddListener(func(event int, ctx interface{}) {
		if event == EventMaxMemory && !once {
			close(captured)
			once = true
		}
	})

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("error during the Serve: error %v", err)
		}
	}()

	time.Sleep(time.Millisecond * 100)

	lastPID := getPID(s)

	req, err := http.NewRequest("GET", "http://localhost:7033", nil)
	assert.NoError(t, err)

	for {
		select {
		case <-captured:
			_, err := http.DefaultClient.Do(req)
			if err != nil {
				c.Stop()
				t.Errorf("error during sending the http request: error %v", err)
			}
			assert.NotEqual(t, lastPID, getPID(s))
			c.Stop()
			return
		default:
			_, err := http.DefaultClient.Do(req)
			if err != nil {
				c.Stop()
				t.Errorf("error during sending the http request: error %v", err)
			}
			c.Stop()
			return
		}
	}
}
func getPID(s interface{}) string {
	if len(s.(*rrhttp.Service).Server().Workers()) > 0 {
		w := s.(*rrhttp.Service).Server().Workers()[0]
		return fmt.Sprintf("%v", *w.Pid)
	} else {
		panic("no workers")
	}
}
