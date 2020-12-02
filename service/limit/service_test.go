package limit

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	json "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	rrhttp "github.com/spiral/roadrunner/service/http"
	"github.com/stretchr/testify/assert"
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
	j := json.ConfigCompatibleWithStandardLibrary
	err := j.Unmarshal([]byte(cfg.target), out)

	if cl, ok := out.(*Config); ok {
		// to speed up tests
		cl.Interval = time.Millisecond
	}

	return err
}

func Test_Service_PidEcho(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(rrhttp.ID, &rrhttp.Service{})
		c.Register(ID, &Service{})

		err := c.Init(&testCfg{
			httpCfg: `{
			"address": ":27029",
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
		})
		if err != nil {
			return err
		}

		s, _ := c.Get(rrhttp.ID)
		assert.NotNil(t, s)

		go func() {
			err := c.Serve()
			if err != nil {
				t.Errorf("error during the Serve: error %v", err)
			}
		}()

		time.Sleep(time.Millisecond * 800)
		req, err := http.NewRequest("GET", "http://localhost:27029", nil)
		if err != nil {
			return err
		}

		r, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}

		assert.Equal(t, getPID(s), string(b))

		err2 := r.Body.Close()
		if err2 != nil {
			t.Errorf("error during the body closing: error %v", err2)
		}
		c.Stop()
		return nil

	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}

}

func Test_Service_ListenerPlusTTL(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(rrhttp.ID, &rrhttp.Service{})
		c.Register(ID, &Service{})

		err := c.Init(&testCfg{
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
		})
		if err != nil {
			return err
		}

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
		if err != nil {
			return err
		}

		r, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		assert.Equal(t, lastPID, string(b))

		<-captured

		// clean state
		req, err = http.NewRequest("GET", "http://localhost:7030?new", nil)
		if err != nil {
			return err
		}

		_, err = http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		assert.NotEqual(t, lastPID, getPID(s))

		c.Stop()

		err2 := r.Body.Close()
		if err2 != nil {
			t.Errorf("error during the body closing: error %v", err2)
		}

		return nil
	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}

}

func Test_Service_ListenerPlusIdleTTL(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(rrhttp.ID, &rrhttp.Service{})
		c.Register(ID, &Service{})

		err := c.Init(&testCfg{
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
		})
		if err != nil {
			return err
		}

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
		if err != nil {
			return err
		}

		r, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		assert.Equal(t, lastPID, string(b))

		<-captured

		// clean state
		req, err = http.NewRequest("GET", "http://localhost:7031?new", nil)
		if err != nil {
			return err
		}

		_, err = http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		assert.NotEqual(t, lastPID, getPID(s))

		c.Stop()
		err2 := r.Body.Close()
		if err2 != nil {
			t.Errorf("error during the body closing: error %v", err2)
		}
		return nil
	}, bkoff)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Service_Listener_MaxExecTTL(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {

		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(rrhttp.ID, &rrhttp.Service{})
		c.Register(ID, &Service{})

		err := c.Init(&testCfg{
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
		})
		if err != nil {
			return err
		}

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
		if err != nil {
			return err
		}

		r, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		assert.Equal(t, 500, r.StatusCode)

		<-captured

		c.Stop()
		return nil
	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}
}

func Test_Service_Listener_MaxMemoryUsage(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(rrhttp.ID, &rrhttp.Service{})
		c.Register(ID, &Service{})

		err := c.Init(&testCfg{
			httpCfg: `{
			"address": ":10033",
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
		})
		if err != nil {
			return err
		}

		time.Sleep(time.Second * 3)
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

		time.Sleep(time.Millisecond * 500)

		lastPID := getPID(s)

		req, err := http.NewRequest("GET", "http://localhost:10033", nil)
		if err != nil {
			return err
		}

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
				return nil
			default:
				_, err := http.DefaultClient.Do(req)
				if err != nil {
					c.Stop()
					t.Errorf("error during sending the http request: error %v", err)
				}
				c.Stop()
				return nil
			}
		}
	}, bkoff)

	if err != nil {
		t.Fatal(err)
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
