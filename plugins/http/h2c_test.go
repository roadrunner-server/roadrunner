package http

import (
	"net/http"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	"github.com/stretchr/testify/assert"
)

func Test_Service_H2C(t *testing.T) {
	bkoff := backoff.NewExponentialBackOff()
	bkoff.MaxElapsedTime = time.Second * 15

	err := backoff.Retry(func() error {
		logger, _ := test.NewNullLogger()
		logger.SetLevel(logrus.DebugLevel)

		c := service.NewContainer(logger)
		c.Register(ID, &Service{})

		err := c.Init(&testCfg{httpCfg: `{
			"address": ":6029",
			"http2": {"h2c":true},
			"workers":{
				"command": "php ../../tests/http/client.php echo pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1
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
				t.Errorf("error serving: %v", err)
			}
		}()
		time.Sleep(time.Millisecond * 100)
		defer c.Stop()

		req, err := http.NewRequest("PRI", "http://localhost:6029?hello=world", nil)
		if err != nil {
			return err
		}

		req.Header.Add("Upgrade", "h2c")
		req.Header.Add("Connection", "HTTP2-Settings")
		req.Header.Add("HTTP2-Settings", "")

		r, err2 := http.DefaultClient.Do(req)
		if err2 != nil {
			return err2
		}

		assert.Equal(t, "101 Switching Protocols", r.Status)

		err3 := r.Body.Close()
		if err3 != nil {
			return err3
		}
		return nil
	}, bkoff)

	if err != nil {
		t.Fatal(err)
	}
}
