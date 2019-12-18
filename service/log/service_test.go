package log

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	rrHttp "github.com/spiral/roadrunner/service/http"
	"github.com/spiral/roadrunner/service/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestService(t *testing.T) {
	c, hook := initContainer()
	c.Register(rrHttp.ID, &rrHttp.Service{})

	err := c.Init(&testCfg{
		rpcCfg: `{"enable":true}`,
		httpCfg: `{
			"address": "localhost:2115",

			"workers":{
				"command": "php ../../tests/http/client.php log pipes",
				"pool": {"numWorkers": 1}
			}
		}`,
	})

	go func() {
		err := c.Serve()

		require.NoError(t, err)
	}()
	defer c.Stop()

	time.Sleep(300 * time.Millisecond)

	assert.NoError(t, err)

	tests := []struct {
		psrLevel    string
		logrusLevel string
	}{
		{"emergency", "error"},
		{"alert", "error"},
		{"critical", "error"},
		{"error", "error"},
		{"warning", "warning"},
		{"notice", "info"},
		{"info", "info"},
		{"debug", "debug"},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("PSR %s level log with %s logrus level", test.psrLevel, test.logrusLevel), func(t *testing.T) {
			hook.Reset()
			res, err := get(fmt.Sprintf("http://localhost:2115/?level=%s", test.psrLevel))

			if assert.NoError(t, err) {
				assert.Equal(t, http.StatusCreated, res.StatusCode)

				e := hook.LastEntry()
				assert.NotNil(t, e)
				assert.Equal(t, test.logrusLevel, e.Level.String())
				assert.Equal(t, test.psrLevel, e.Data["psr_level"])
			}
		})
	}

	t.Run("With fields", func(t *testing.T) {
		hook.Reset()
		res, err := get("http://localhost:2115/?level=info&message=hello&fields[foo]=bar")

		if assert.NoError(t, err) {
			assert.Equal(t, http.StatusCreated, res.StatusCode)

			e := hook.LastEntry()
			assert.NotNil(t, e)
			assert.Equal(t, logrus.InfoLevel, e.Level)
			assert.Equal(t, "hello", e.Message)
			assert.Equal(t, "bar", e.Data["foo"])
		}
	})
}

func initContainer() (service.Container, *test.Hook) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)

	c := service.NewContainer(logger)
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(ID, &Service{})

	return c, hook
}

type testCfg struct {
	rpcCfg  string
	httpCfg string
	target  string
}

func (cfg *testCfg) Get(name string) service.Config {

	if name == rpc.ID {
		return &testCfg{target: cfg.rpcCfg}
	}
	if name == rrHttp.ID {
		return &testCfg{target: cfg.httpCfg}
	}

	return nil
}

func (cfg *testCfg) Unmarshal(out interface{}) error {
	err := json.Unmarshal([]byte(cfg.target), out)
	return err
}

// get request and return body
func get(url string) (*http.Response, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return r, err
}
