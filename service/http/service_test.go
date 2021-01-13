package http

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/service/env"
	"github.com/spiral/roadrunner/service/rpc"
	"github.com/stretchr/testify/assert"
)

type testCfg struct {
	httpCfg string
	rpcCfg  string
	envCfg  string
	target  string
}

func (cfg *testCfg) Get(name string) service.Config {
	if name == ID {
		if cfg.httpCfg == "" {
			return nil
		}

		return &testCfg{target: cfg.httpCfg}
	}

	if name == rpc.ID {
		return &testCfg{target: cfg.rpcCfg}
	}

	if name == env.ID {
		return &testCfg{target: cfg.envCfg}
	}

	return nil
}
func (cfg *testCfg) Unmarshal(out interface{}) error {
	return json.Unmarshal([]byte(cfg.target), out)
}

func Test_Service_NoConfig(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	err := c.Init(&testCfg{httpCfg: `{"Enable":true}`})
	assert.Error(t, err)

	s, st := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusInactive, st)
}

func Test_Service_Configure_Disable(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{}))

	s, st := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusInactive, st)
}

func Test_Service_Configure_Enable(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(ID, &Service{})

		err := c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": ":8070",
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
	}`})
		if err != nil {
			return err
		}

		s, st := c.Get(ID)
		assert.NotNil(t, s)
		assert.Equal(t, service.StatusOK, st)

		return nil
	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}

}

func Test_Service_Echo(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(ID, &Service{})

		err := c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": ":6536",
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
	}`})
		if err != nil {
			return err
		}

		s, st := c.Get(ID)
		assert.NotNil(t, s)
		assert.Equal(t, service.StatusOK, st)

		// should do nothing
		s.(*Service).Stop()

		go func() {
			err := c.Serve()
			if err != nil {
				t.Errorf("serve error: %v", err)
			}
		}()

		time.Sleep(time.Millisecond * 100)

		req, err := http.NewRequest("GET", "http://localhost:6536?hello=world", nil)
		if err != nil {
			c.Stop()
			return err
		}

		r, err := http.DefaultClient.Do(req)
		if err != nil {
			c.Stop()
			return err
		}
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			c.Stop()
			return err
		}
		assert.Equal(t, 201, r.StatusCode)
		assert.Equal(t, "WORLD", string(b))

		err = r.Body.Close()
		if err != nil {
			c.Stop()
			return err
		}

		c.Stop()
		return nil
	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}
}

func Test_Service_Env(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(env.ID, env.NewService(map[string]string{"rr": "test"}))
		c.Register(ID, &Service{})

		err := c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": ":10031",
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../tests/http/client.php env pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`, envCfg: `{"env_key":"ENV_VALUE"}`})
		if err != nil {
			return err
		}

		s, st := c.Get(ID)
		assert.NotNil(t, s)
		assert.Equal(t, service.StatusOK, st)

		// should do nothing
		s.(*Service).Stop()

		go func() {
			err := c.Serve()
			if err != nil {
				t.Errorf("serve error: %v", err)
			}
		}()

		time.Sleep(time.Millisecond * 500)

		req, err := http.NewRequest("GET", "http://localhost:10031", nil)
		if err != nil {
			c.Stop()
			return err
		}

		r, err := http.DefaultClient.Do(req)
		if err != nil {
			c.Stop()
			return err
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			c.Stop()
			return err
		}

		assert.Equal(t, 200, r.StatusCode)
		assert.Equal(t, "ENV_VALUE", string(b))

		err = r.Body.Close()
		if err != nil {
			c.Stop()
			return err
		}

		c.Stop()
		return nil
	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}

}

func Test_Service_ErrorEcho(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(ID, &Service{})

		err := c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": ":6030",
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../tests/http/client.php echoerr pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`})
		if err != nil {
			return err
		}

		s, st := c.Get(ID)
		assert.NotNil(t, s)
		assert.Equal(t, service.StatusOK, st)

		goterr := make(chan interface{})
		s.(*Service).AddListener(func(event int, ctx interface{}) {
			if event == roadrunner.EventStderrOutput {
				if string(ctx.([]byte)) == "WORLD\n" {
					goterr <- nil
				}
			}
		})

		go func() {
			err := c.Serve()
			if err != nil {
				t.Errorf("serve error: %v", err)
			}
		}()

		time.Sleep(time.Millisecond * 500)

		req, err := http.NewRequest("GET", "http://localhost:6030?hello=world", nil)
		if err != nil {
			c.Stop()
			return err
		}

		r, err := http.DefaultClient.Do(req)
		if err != nil {
			c.Stop()
			return err
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			c.Stop()
			return err
		}

		<-goterr

		assert.Equal(t, 201, r.StatusCode)
		assert.Equal(t, "WORLD", string(b))
		err = r.Body.Close()
		if err != nil {
			c.Stop()
			return err
		}

		c.Stop()

		return nil
	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}
}

func Test_Service_Middleware(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(ID, &Service{})

		err := c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": ":6032",
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
	}`})
		if err != nil {
			return err
		}

		s, st := c.Get(ID)
		assert.NotNil(t, s)
		assert.Equal(t, service.StatusOK, st)

		s.(*Service).AddMiddleware(func(f http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/halt" {
					w.WriteHeader(500)
					_, err := w.Write([]byte("halted"))
					if err != nil {
						t.Errorf("error writing the data to the http reply: error %v", err)
					}
				} else {
					f(w, r)
				}
			}
		})

		go func() {
			err := c.Serve()
			if err != nil {
				t.Errorf("serve error: %v", err)
			}
		}()
		time.Sleep(time.Millisecond * 500)

		req, err := http.NewRequest("GET", "http://localhost:6032?hello=world", nil)
		if err != nil {
			c.Stop()
			return err
		}

		r, err := http.DefaultClient.Do(req)
		if err != nil {
			c.Stop()
			return err
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			c.Stop()
			return err
		}

		assert.Equal(t, 201, r.StatusCode)
		assert.Equal(t, "WORLD", string(b))

		err = r.Body.Close()
		if err != nil {
			c.Stop()
			return err
		}

		req, err = http.NewRequest("GET", "http://localhost:6032/halt", nil)
		if err != nil {
			c.Stop()
			return err
		}

		r, err = http.DefaultClient.Do(req)
		if err != nil {
			c.Stop()
			return err
		}
		b, err = ioutil.ReadAll(r.Body)
		if err != nil {
			c.Stop()
			return err
		}

		assert.Equal(t, 500, r.StatusCode)
		assert.Equal(t, "halted", string(b))

		err = r.Body.Close()
		if err != nil {
			c.Stop()
			return err
		}
		c.Stop()

		return nil
	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}

}

func Test_Service_Listener(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(ID, &Service{})

		err := c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": ":6033",
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
	}`})
		if err != nil {
			return err
		}

		s, st := c.Get(ID)
		assert.NotNil(t, s)
		assert.Equal(t, service.StatusOK, st)

		stop := make(chan interface{})
		s.(*Service).AddListener(func(event int, ctx interface{}) {
			if event == roadrunner.EventServerStart {
				stop <- nil
			}
		})

		go func() {
			err := c.Serve()
			if err != nil {
				t.Errorf("serve error: %v", err)
			}
		}()
		time.Sleep(time.Millisecond * 500)

		c.Stop()
		assert.True(t, true)

		return nil
	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}
}

func Test_Service_Error(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(ID, &Service{})

		err := c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": ":6034",
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../tests/http/client.php echo pipes",
				"relay": "---",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`})
		if err != nil {
			return err
		}

		// assert error
		err = c.Serve()
		if err == nil {
			return err
		}

		return nil
	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}
}

func Test_Service_Error2(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(ID, &Service{})

		err := c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": ":6035",
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../tests/http/client.php broken pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`})
		if err != nil {
			return err
		}

		// assert error
		err = c.Serve()
		if err == nil {
			return err
		}

		return nil
	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}
}

func Test_Service_Error3(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(ID, &Service{})

		err := c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": ":6036",
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers"
				"command": "php ../../tests/http/client.php broken pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`})
		// assert error
		if err == nil {
			return err
		}

		return nil
	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}

}

func Test_Service_Error4(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(ID, &Service{})

		err := c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": "----",
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../tests/http/client.php broken pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`})
		// assert error
		if err != nil {
			return nil
		}

		return err
	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}
}

func tmpDir() string {
	p := os.TempDir()
	r, _ := json.Marshal(p)

	return string(r)
}
