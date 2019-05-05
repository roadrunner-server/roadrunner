package static

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	rrhttp "github.com/spiral/roadrunner/service/http"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

type testCfg struct {
	httpCfg string
	static  string
	target  string
}

func (cfg *testCfg) Get(name string) service.Config {
	if name == rrhttp.ID {
		return &testCfg{target: cfg.httpCfg}
	}

	if name == ID {
		return &testCfg{target: cfg.static}
	}
	return nil
}
func (cfg *testCfg) Unmarshal(out interface{}) error {
	return json.Unmarshal([]byte(cfg.target), out)
}

func get(url string) (string, *http.Response, error) {
	r, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	return string(b), r, err
}

func Test_Files(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		static: `{"enable":true, "dir":"../../tests", "forbid":[]}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../tests/http/client.php pid pipes",
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

	b, _, _ := get("http://localhost:6029/sample.txt")
	assert.Equal(t, "sample", b)
}

func Test_Disabled(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		static: `{"enable":true, "dir":"../../tests", "forbid":[]}`,
	}))

	s, st := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusInactive, st)
}

func Test_Files_Disable(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		static: `{"enable":false, "dir":"../../tests", "forbid":[".php"]}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../tests/http/client.php echo pipes",
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

	b, _, _ := get("http://localhost:6029/client.php?hello=world")
	assert.Equal(t, "WORLD", b)
}

func Test_Files_Error(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.Error(t, c.Init(&testCfg{
		static: `{"enable":true, "dir":"dir/invalid", "forbid":[".php"]}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../tests/http/client.php echo pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1,
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000
				}
			}
	}`}))
}

func Test_Files_Error2(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.Error(t, c.Init(&testCfg{
		static: `{"enable":true, "dir":"dir/invalid", "forbid":[".php"]`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../tests/http/client.php echo pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1,
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000
				}
			}
	}`}))
}

func Test_Files_Forbid(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		static: `{"enable":true, "dir":"../../tests", "forbid":[".php"]}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../tests/http/client.php echo pipes",
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

	b, _, _ := get("http://localhost:6029/client.php?hello=world")
	assert.Equal(t, "WORLD", b)
}

func Test_Files_Always(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		static: `{"enable":true, "dir":"../../tests", "forbid":[".php"], "always":[".ico"]}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../tests/http/client.php echo pipes",
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

	_, r, _ := get("http://localhost:6029/favicon.ico")
	assert.Equal(t, 404, r.StatusCode)
}

func Test_Files_NotFound(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		static: `{"enable":true, "dir":"../../tests", "forbid":[".php"]}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../tests/http/client.php echo pipes",
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

	b, _, _ := get("http://localhost:6029/client.XXX?hello=world")
	assert.Equal(t, "WORLD", b)
}

func Test_Files_Dir(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		static: `{"enable":true, "dir":"../../tests", "forbid":[".php"]}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../tests/http/client.php echo pipes",
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

	b, _, _ := get("http://localhost:6029/http?hello=world")
	assert.Equal(t, "WORLD", b)
}

func Test_Files_NotForbid(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		static: `{"enable":true, "dir":"../../tests", "forbid":[]}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../tests/http/client.php pid pipes",
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

	b, _, _ := get("http://localhost:6029/client.php")
	assert.Equal(t, all("../../tests/client.php"), b)
}

func tmpDir() string {
	p := os.TempDir()
	r, _ := json.Marshal(p)

	return string(r)
}

func all(fn string) string {
	f, _ := os.Open(fn)
	defer f.Close()

	b := &bytes.Buffer{}
	io.Copy(b, f)

	return b.String()
}
