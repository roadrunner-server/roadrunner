package http

import (
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func Test_Service_H2C(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{
			"address": ":6029",
			"http2": {"h2c":true},
			"workers":{
				"command": "php ../../tests/http/client.php echo pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1
				}
			}
	}`}))

	s, st := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusOK, st)

	// should do nothing
	s.(*Service).Stop()

	go func() { c.Serve() }()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	req, err := http.NewRequest("PRI", "http://localhost:6029?hello=world", nil)
	assert.NoError(t, err)

	req.Header.Add("Upgrade", "h2c")
	req.Header.Add("Connection", "HTTP2-Settings")
	req.Header.Add("HTTP2-Settings", "")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	assert.Equal(t, "101 Switching Protocols", r.Status)

	// will fail with h2c notice
}
