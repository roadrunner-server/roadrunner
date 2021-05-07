package static

import (
	"bytes"
	json "github.com/json-iterator/go"
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
	j := json.ConfigCompatibleWithStandardLibrary
	return j.Unmarshal([]byte(cfg.target), out)
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
			"address": ":8029",
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

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("serve error: %v", err)
		}
	}()

	time.Sleep(time.Second)


	b, _, _ := get("http://localhost:8029/sample.txt")
	assert.Equal(t, "sample", b)
	c.Stop()
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
			"address": ":8030",
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

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("serve error: %v", err)
		}
	}()

	time.Sleep(time.Second)

	b, _, err := get("http://localhost:8030/client.php?hello=world")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "WORLD", b)
	c.Stop()
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
			"address": ":8031",
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
			"address": ":8032",
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
			"address": ":8033",
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

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("serve error: %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 500)

	b, _, err := get("http://localhost:8033/client.php?hello=world")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "WORLD", b)
	c.Stop()
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
			"address": ":8034",
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

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("serve error: %v", err)
		}
	}()

	time.Sleep(time.Millisecond * 500)

	_, r, err := get("http://localhost:8034/favicon.ico")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 404, r.StatusCode)
	c.Stop()
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
			"address": ":8035",
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

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("serve error: %v", err)
		}
	}()

	time.Sleep(time.Millisecond * 500)

	b, _, _ := get("http://localhost:8035/client.XXX?hello=world")
	assert.Equal(t, "WORLD", b)
	c.Stop()
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
			"address": ":8036",
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

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("serve error: %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 500)

	b, _, _ := get("http://localhost:8036/http?hello=world")
	assert.Equal(t, "WORLD", b)
	c.Stop()
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
			"address": ":8037",
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

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("serve error: %v", err)
		}
	}()

	time.Sleep(time.Millisecond * 500)

	b, _, _ := get("http://localhost:8037/client.php")
	assert.Equal(t, all("../../tests/client.php"), b)
	assert.Equal(t, all("../../tests/client.php"), b)
	c.Stop()
}

func TestStatic_Headers(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		static: `{"enable":true, "dir":"../../tests", "forbid":[], "request":{"input": "custom-header"}, "response":{"output": "output-header"}}`,
		httpCfg: `{
			"enable": true,
			"address": ":8037",
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

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("serve error: %v", err)
		}
	}()

	time.Sleep(time.Millisecond * 500)

	req, err := http.NewRequest("GET", "http://localhost:8037/client.php", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Header.Get("Output") != "output-header" {
		t.Fatal("can't find output header in response")
	}


	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, all("../../tests/client.php"), string(b))
	assert.Equal(t, all("../../tests/client.php"), string(b))
	c.Stop()
}

func get(url string) (string, *http.Response, error) {
	r, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", nil, err
	}

	err = r.Body.Close()
	if err != nil {
		return "", nil, err
	}

	return string(b), r, err
}

func tmpDir() string {
	p := os.TempDir()
	j := json.ConfigCompatibleWithStandardLibrary
	r, _ := j.Marshal(p)

	return string(r)
}

func all(fn string) string {
	f, _ := os.Open(fn)

	b := &bytes.Buffer{}
	_, err := io.Copy(b, f)
	if err != nil {
		return ""
	}

	err = f.Close()
	if err != nil {
		return ""
	}

	return b.String()
}
