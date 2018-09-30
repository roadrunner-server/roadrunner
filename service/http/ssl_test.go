package http

import (
	"crypto/tls"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

var sslClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func Test_SSL_Service_Echo(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{
			"address": ":6029",
			"ssl": {
				"port": 6900,
				"key": "fixtures/server.key",
				"cert": "fixtures/server.crt"
			},
			"workers":{
				"command": "php ../../tests/http/client.php echo pipes",
				"pool": {"numWorkers": 1}
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

	req, err := http.NewRequest("GET", "https://localhost:6900?hello=world", nil)
	assert.NoError(t, err)

	r, err := sslClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))
}

func Test_SSL_Service_NoRedirect(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{
			"address": ":6029",
			"ssl": {
				"port": 6900,
				"key": "fixtures/server.key",
				"cert": "fixtures/server.crt"
			},
			"workers":{
				"command": "php ../../tests/http/client.php echo pipes",
				"pool": {"numWorkers": 1}
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

	req, err := http.NewRequest("GET", "http://localhost:6029?hello=world", nil)
	assert.NoError(t, err)

	r, err := sslClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	assert.Nil(t, r.TLS)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))
}

func Test_SSL_Service_Redirect(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{
			"address": ":6029",
			"ssl": {
				"port": 6900,
				"redirect": true,
				"key": "fixtures/server.key",
				"cert": "fixtures/server.crt"
			},
			"workers":{
				"command": "php ../../tests/http/client.php echo pipes",
				"pool": {"numWorkers": 1}
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

	req, err := http.NewRequest("GET", "http://localhost:6029?hello=world", nil)
	assert.NoError(t, err)

	r, err := sslClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	assert.NotNil(t, r.TLS)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))
}

func Test_SSL_Service_Push(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{
			"address": ":6029",
			"ssl": {
				"port": 6900,
				"redirect": true,
				"key": "fixtures/server.key",
				"cert": "fixtures/server.crt"
			},
			"workers":{
				"command": "php ../../tests/http/client.php push pipes",
				"pool": {"numWorkers": 1}
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

	req, err := http.NewRequest("GET", "https://localhost:6900?hello=world", nil)
	assert.NoError(t, err)

	r, err := sslClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	assert.NotNil(t, r.TLS)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, "", r.Header.Get("http2-push"))

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))
}
