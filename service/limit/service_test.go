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
			"address": ":6029",
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

	go func() { c.Serve() }()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	req, err := http.NewRequest("GET", "http://localhost:6029", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, getPID(s), string(b))
}

func getPID(s interface{}) string {
	w := s.(*rrhttp.Service).Server().Workers()[0]
	return fmt.Sprintf("%v", *w.Pid)
}
