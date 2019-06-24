package http

import (
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	"github.com/stretchr/testify/assert"
	"github.com/yookoala/gofast"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"
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
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

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
}
