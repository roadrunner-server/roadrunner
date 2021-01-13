package headers

import (
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
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(rrhttp.ID, &rrhttp.Service{})
		c.Register(ID, &Service{})

		assert.NoError(t, c.Init(&testCfg{
			headers: `{"request":{"input": "custom-header"}}`,
			httpCfg: `{
			"enable": true,
			"address": ":6078",
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

		go func() {
			err := c.Serve()
			if err != nil {
				t.Errorf("error during Serve: error %v", err)
			}
		}()

		time.Sleep(time.Millisecond * 100)
		defer c.Stop()

		req, err := http.NewRequest("GET", "http://localhost:6078?hello=value", nil)
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

		assert.Equal(t, 200, r.StatusCode)
		assert.Equal(t, "CUSTOM-HEADER", string(b))

		err = r.Body.Close()
		if err != nil {
			return err
		}

		return nil
	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}
}

func Test_ResponseHeaders(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(rrhttp.ID, &rrhttp.Service{})
		c.Register(ID, &Service{})

		assert.NoError(t, c.Init(&testCfg{
			headers: `{"response":{"output": "output-header"},"request":{"input": "custom-header"}}`,
			httpCfg: `{
			"enable": true,
			"address": ":6079",
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

		go func() {
			err := c.Serve()
			if err != nil {
				t.Errorf("error during the Serve: error %v", err)
			}
		}()
		time.Sleep(time.Millisecond * 100)
		defer c.Stop()

		req, err := http.NewRequest("GET", "http://localhost:6079?hello=value", nil)
		if err != nil {
			return err
		}

		r, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		assert.Equal(t, "output-header", r.Header.Get("output"))

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		assert.Equal(t, 200, r.StatusCode)
		assert.Equal(t, "CUSTOM-HEADER", string(b))

		err = r.Body.Close()
		if err != nil {
			return err
		}

		return nil
	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}
}

func TestCORS_OPTIONS(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
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
			"address": ":16379",
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

		go func() {
			err := c.Serve()
			if err != nil {
				t.Errorf("error during the Serve: error %v", err)
			}
		}()
		time.Sleep(time.Millisecond * 100)
		defer c.Stop()

		req, err := http.NewRequest("OPTIONS", "http://localhost:16379", nil)
		if err != nil {
			return err
		}

		r, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		assert.Equal(t, "true", r.Header.Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "*", r.Header.Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "GET,POST,PUT,DELETE", r.Header.Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "*", r.Header.Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "600", r.Header.Get("Access-Control-Max-Age"))
		assert.Equal(t, "true", r.Header.Get("Access-Control-Allow-Credentials"))

		_, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		assert.Equal(t, 200, r.StatusCode)

		err = r.Body.Close()
		if err != nil {
			return err
		}

		return nil
	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}
}

func TestCORS_Pass(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
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
			"address": ":6672",
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

		go func() {
			err := c.Serve()
			if err != nil {
				t.Errorf("error during the Serve: error %v", err)
			}
		}()
		time.Sleep(time.Millisecond * 100)
		defer c.Stop()

		req, err := http.NewRequest("GET", "http://localhost:6672", nil)
		if err != nil {
			return err
		}

		r, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		assert.Equal(t, "true", r.Header.Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "*", r.Header.Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "*", r.Header.Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", r.Header.Get("Access-Control-Allow-Credentials"))

		_, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		assert.Equal(t, 200, r.StatusCode)

		err = r.Body.Close()
		if err != nil {
			return err
		}

		return nil
	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}
}
