package health

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	json "github.com/json-iterator/go"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	rrhttp "github.com/spiral/roadrunner/service/http"
	"github.com/stretchr/testify/assert"
)

type testCfg struct {
	healthCfg string
	httpCfg   string
	target    string
}

func (cfg *testCfg) Get(name string) service.Config {
	if name == ID {
		return &testCfg{target: cfg.healthCfg}
	}

	if name == rrhttp.ID {
		return &testCfg{target: cfg.httpCfg}
	}

	return nil
}

func (cfg *testCfg) Unmarshal(out interface{}) error {
	j := json.ConfigCompatibleWithStandardLibrary
	err := j.Unmarshal([]byte(cfg.target), out)
	return err
}

func TestService_Serve(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		healthCfg: `{
			"address": "localhost:2116"
		}`,
		httpCfg: `{
			"address": "localhost:2115",
			"workers":{
				"command": "php ../../tests/http/client.php echo pipes",
				"pool": {"numWorkers": 1}
			}
		}`,
	}))

	s, status := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusOK, status)

	hS, httpStatus := c.Get(rrhttp.ID)
	assert.NotNil(t, hS)
	assert.Equal(t, service.StatusOK, httpStatus)

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("serve error: %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 500)
	defer c.Stop()

	_, res, err := get("http://localhost:2116/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestService_Serve_DeadWorker(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		healthCfg: `{
			"address": "localhost:2117"
		}`,
		httpCfg: `{
			"address": "localhost:2118",
			"workers":{
				"command": "php ../../tests/http/slow-client.php echo pipes 1000",
				"pool": {"numWorkers": 1}
			}
		}`,
	}))

	s, status := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusOK, status)

	hS, httpStatus := c.Get(rrhttp.ID)
	assert.NotNil(t, hS)
	assert.Equal(t, service.StatusOK, httpStatus)

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("server error: %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 500)
	defer c.Stop()

	// Kill the worker
	httpSvc := hS.(*rrhttp.Service)
	err := httpSvc.Server().Workers()[0].Kill()
	if err != nil {
		t.Errorf("error killing the worker: error %v", err)
	}

	// Check health check
	_, res, err := get("http://localhost:2117/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestService_Serve_DeadWorkerStillHealthy(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		healthCfg: `{
			"address": "localhost:2119"
		}`,
		httpCfg: `{
			"address": "localhost:2120",
			"workers":{
				"command": "php ../../tests/http/client.php echo pipes",
				"pool": {"numWorkers": 2}
			}
		}`,
	}))

	s, status := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusOK, status)

	hS, httpStatus := c.Get(rrhttp.ID)
	assert.NotNil(t, hS)
	assert.Equal(t, service.StatusOK, httpStatus)

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("serve error: %v", err)
		}
	}()
	time.Sleep(time.Second * 1)
	defer c.Stop()

	// Kill one of the workers
	httpSvc := hS.(*rrhttp.Service)
	err := httpSvc.Server().Workers()[0].Kill()
	if err != nil {
		t.Errorf("error killing the worker: error %v", err)
	}

	// Check health check
	_, res, err := get("http://localhost:2119/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestService_Serve_NoHTTPService(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		healthCfg: `{
			"address": "localhost:2121"
		}`,
	}))

	s, status := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusInactive, status)
}

func TestService_Serve_NoServer(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	healthSvc := &Service{}

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, &rrhttp.Service{})
	c.Register(ID, healthSvc)

	assert.NoError(t, c.Init(&testCfg{
		healthCfg: `{
			"address": "localhost:2122"
		}`,
		httpCfg: `{
			"address": "localhost:2123",
			"workers":{
				"command": "php ../../tests/http/client.php echo pipes",
				"pool": {"numWorkers": 1}
			}
		}`,
	}))

	s, status := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusOK, status)

	hS, httpStatus := c.Get(rrhttp.ID)
	assert.NotNil(t, hS)
	assert.Equal(t, service.StatusOK, httpStatus)

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("serve error: %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 500)
	defer c.Stop()

	// Set the httpService to nil
	healthSvc.httpService = nil

	_, res, err := get("http://localhost:2122/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestService_Serve_NoPool(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	httpSvc := &rrhttp.Service{}

	c := service.NewContainer(logger)
	c.Register(rrhttp.ID, httpSvc)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		healthCfg: `{
			"address": "localhost:2124"
		}`,
		httpCfg: `{
			"address": "localhost:2125",
			"workers":{
				"command": "php ../../tests/http/client.php echo pipes",
				"pool": {"numWorkers": 1}
			}
		}`,
	}))

	s, status := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusOK, status)

	hS, httpStatus := c.Get(rrhttp.ID)
	assert.NotNil(t, hS)
	assert.Equal(t, service.StatusOK, httpStatus)

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("serve error: %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 500)
	defer c.Stop()

	// Stop the pool
	httpSvc.Server().Stop()

	_, res, err := get("http://localhost:2124/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

// get request and return body
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
