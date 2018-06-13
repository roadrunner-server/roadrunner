package http

import (
	"testing"
	"github.com/spiral/roadrunner/service"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/sirupsen/logrus"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"os"
	"time"
	"net/http"
	"io/ioutil"
)

type testCfg struct{ httpCfg string }

func (cfg *testCfg) Get(name string) service.Config {
	if name == Name {
		return &testCfg{cfg.httpCfg}
	}
	return nil
}
func (cfg *testCfg) Unmarshal(out interface{}) error {
	json.Unmarshal([]byte(cfg.httpCfg), out)
	return nil
}

func Test_Service_NoConfig(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(Name, &Service{})

	assert.NoError(t, c.Init(&testCfg{`{}`}))

	s, st := c.Get(Name)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusRegistered, st)
}

func Test_Service_Configure_Disable(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(Name, &Service{})

	assert.NoError(t, c.Init(&testCfg{`{
			"enable": false,
			"address": ":8070",
			"maxRequest": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../php-src/tests/http/client.php echo pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))

	s, st := c.Get(Name)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusRegistered, st)
}

func Test_Service_Configure_Enable(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(Name, &Service{})

	assert.NoError(t, c.Init(&testCfg{`{
			"enable": true,
			"address": ":8070",
			"maxRequest": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../php-src/tests/http/client.php echo pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))

	s, st := c.Get(Name)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusConfigured, st)
}

func Test_Service_Echo(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(Name, &Service{})

	assert.NoError(t, c.Init(&testCfg{`{
			"enable": true,
			"address": ":6029",
			"maxRequest": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../php-src/tests/http/client.php echo pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))

	s, st := c.Get(Name)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusConfigured, st)

	go func() { c.Serve() }()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	req, err := http.NewRequest("GET", "http://localhost:6029?hello=world", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))
}

func tmpDir() string {
	p := os.TempDir()
	r, _ := json.Marshal(p)

	return string(r)
}
