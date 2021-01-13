package http

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	"github.com/stretchr/testify/assert"
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

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("error during the Serve: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 500)

	req, err := http.NewRequest("GET", "https://localhost:6900?hello=world", nil)
	assert.NoError(t, err)

	r, err := sslClient.Do(req)
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err2 := r.Body.Close()
	if err2 != nil {
		t.Errorf("fail to close the Body: error %v", err2)
	}

	c.Stop()
}

func Test_SSL_Service_NoRedirect(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{
			"address": ":6030",
			"ssl": {
				"port": 6901,
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

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("error during the Serve: error %v", err)
		}
	}()

	time.Sleep(time.Millisecond * 500)

	req, err := http.NewRequest("GET", "http://localhost:6030?hello=world", nil)
	assert.NoError(t, err)

	r, err := sslClient.Do(req)
	assert.NoError(t, err)

	assert.Nil(t, r.TLS)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err2 := r.Body.Close()
	if err2 != nil {
		t.Errorf("fail to close the Body: error %v", err2)
	}
	c.Stop()
}

func Test_SSL_Service_Redirect(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{
			"address": ":6831",
			"ssl": {
				"port": 6902,
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

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("error during the Serve: error %v", err)
		}
	}()

	time.Sleep(time.Millisecond * 500)

	req, err := http.NewRequest("GET", "http://localhost:6831?hello=world", nil)
	assert.NoError(t, err)

	r, err := sslClient.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, r.TLS)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err2 := r.Body.Close()
	if err2 != nil {
		t.Errorf("fail to close the Body: error %v", err2)
	}
	c.Stop()
}

func Test_SSL_Service_Push(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{
			"address": ":6032",
			"ssl": {
				"port": 6903,
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

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("error during the Serve: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 500)

	req, err := http.NewRequest("GET", "https://localhost:6903?hello=world", nil)
	assert.NoError(t, err)

	r, err := sslClient.Do(req)
	assert.NoError(t, err)

	assert.NotNil(t, r.TLS)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, "", r.Header.Get("Http2-Push"))

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err2 := r.Body.Close()
	if err2 != nil {
		t.Errorf("fail to close the Body: error %v", err2)
	}
	c.Stop()
}
