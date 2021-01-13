package http

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	"github.com/stretchr/testify/assert"
	"github.com/yookoala/gofast"
)

func Test_FCGI_Service_Echo(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{
			"fcgi": {
				"address": "tcp://0.0.0.0:6082"
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

	go func() { assert.NoError(t, c.Serve()) }()
	time.Sleep(time.Second * 1)

	fcgiConnFactory := gofast.SimpleConnFactory("tcp", "0.0.0.0:6082")

	fcgiHandler := gofast.NewHandler(
		gofast.BasicParamsMap(gofast.BasicSession),
		gofast.SimpleClientFactory(fcgiConnFactory, 0),
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://site.local/?hello=world", nil)
	fcgiHandler.ServeHTTP(w, req)

	body, err := ioutil.ReadAll(w.Result().Body)

	assert.NoError(t, err)
	assert.Equal(t, 201, w.Result().StatusCode)
	assert.Equal(t, "WORLD", string(body))
	c.Stop()
}

func Test_FCGI_Service_Request_Uri(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{
			"fcgi": {
				"address": "tcp://0.0.0.0:6083"
			},
			"workers":{
				"command": "php ../../tests/http/client.php request-uri pipes",
				"pool": {"numWorkers": 1}
			}
	}`}))

	s, st := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusOK, st)

	// should do nothing
	s.(*Service).Stop()

	go func() { assert.NoError(t, c.Serve()) }()
	time.Sleep(time.Second * 1)

	fcgiConnFactory := gofast.SimpleConnFactory("tcp", "0.0.0.0:6083")

	fcgiHandler := gofast.NewHandler(
		gofast.BasicParamsMap(gofast.BasicSession),
		gofast.SimpleClientFactory(fcgiConnFactory, 0),
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://site.local/hello-world", nil)
	fcgiHandler.ServeHTTP(w, req)

	body, err := ioutil.ReadAll(w.Result().Body)

	assert.NoError(t, err)
	assert.Equal(t, 200, w.Result().StatusCode)
	assert.Equal(t, "http://site.local/hello-world", string(body))
	c.Stop()
}
