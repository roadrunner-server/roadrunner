package headers

import (
	"encoding/json"
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
	httpCfg string
	headers string
	target  string
}

func (cfg *testCfg) Get(name string) service.Config {
	if name == rrhttp.ID {
		return &testCfg{target: cfg.httpCfg}
	}

	if name == ID {
		return &testCfg{target: cfg.headers}
	}
	return nil
}

func (cfg *testCfg) Unmarshal(out interface{}) error {
	return json.Unmarshal([]byte(cfg.target), out)
}

func Test_RequestHeaders(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		headers: `{"request":{"input": "custom-header"}}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequestSize": 1024,
			"workers":{
				"command": "php ../../tests/http/client.php header pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))

	go func() { c.Serve() }()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	req, err := http.NewRequest("GET", "http://localhost:6029?hello=value", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "CUSTOM-HEADER", string(b))
}

func Test_ResponseHeaders(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		headers: `{"response":{"output": "output-header"},"request":{"input": "custom-header"}}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequestSize": 1024,
			"workers":{
				"command": "php ../../tests/http/client.php header pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))

	go func() { c.Serve() }()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	req, err := http.NewRequest("GET", "http://localhost:6029?hello=value", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	assert.Equal(t, "output-header", r.Header.Get("output"))

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "CUSTOM-HEADER", string(b))
}

func TestCORS_OPTIONS(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		headers: `{
"cors":{
    "allowedOrigin": "*",
    "allowedHeaders": "*",
    "allowedMethods": "GET,POST,PUT,DELETE",
    "allowCredentials": true,
    "exposedHeaders": "Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma",
    "maxAge": 600
}
}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequestSize": 1024,
			"workers":{
				"command": "php ../../tests/http/client.php headers pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))

	go func() { c.Serve() }()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	req, err := http.NewRequest("OPTIONS", "http://localhost:6029", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	assert.Equal(t, "true", r.Header.Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "*", r.Header.Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "GET,POST,PUT,DELETE", r.Header.Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "*", r.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "600", r.Header.Get("Access-Control-Max-Age"))
	assert.Equal(t, "true", r.Header.Get("Access-Control-Allow-Credentials"))

	_, err = ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, 200, r.StatusCode)
}

func TestCORS_Pass(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		headers: `{
"cors":{
    "allowedOrigin": "*",
    "allowedHeaders": "*",
    "allowedMethods": "GET,POST,PUT,DELETE",
    "allowCredentials": true,
    "exposedHeaders": "Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma",
    "maxAge": 600
}
}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequestSize": 1024,
			"workers":{
				"command": "php ../../tests/http/client.php headers pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))

	go func() { c.Serve() }()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	req, err := http.NewRequest("GET", "http://localhost:6029", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	assert.Equal(t, "true", r.Header.Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "*", r.Header.Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "*", r.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", r.Header.Get("Access-Control-Allow-Credentials"))

	_, err = ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, 200, r.StatusCode)
}
