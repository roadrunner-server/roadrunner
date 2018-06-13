package static

import (
	"testing"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/sirupsen/logrus"
	"github.com/spiral/roadrunner/service"
	"github.com/stretchr/testify/assert"
	"time"
	rrhttp "github.com/spiral/roadrunner/service/http"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"bytes"
	"io"
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
		static: `{"enable":true, "dir":"../../php-src/tests/http", "forbid":[]}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequest": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../php-src/tests/http/client.php pid pipes",
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

func Test_Files_Forbid(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		static: `{"enable":true, "dir":"../../php-src/tests/http", "forbid":[".php"]}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequest": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../php-src/tests/http/client.php pid pipes",
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
	assert.NotEqual(t, "sample", b)
}

func Test_Files_NotForbid(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		static: `{"enable":true, "dir":"../../php-src/tests/http", "forbid":[]}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequest": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../php-src/tests/http/client.php pid pipes",
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
	assert.NotEqual(t, all("../../php-src/tests/http/client.php")+"s", b)
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
