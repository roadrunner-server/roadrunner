package http

import (
	"testing"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/sirupsen/logrus"
	"github.com/spiral/roadrunner/service"
	"github.com/stretchr/testify/assert"
	"time"
	"github.com/spiral/roadrunner/service/rpc"
	"strconv"
)

func Test_RPC(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		rpcCfg: `{"enable":true, "listen":"tcp://:5004"}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequest": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../php-src/tests/http/client.php pid pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))

	s, _ := c.Get(ID)
	ss := s.(*Service)

	s2, _ := c.Get(rpc.ID)
	rs := s2.(*rpc.Service)

	go func() { c.Serve() }()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	res, _, _ := get("http://localhost:6029")
	assert.Equal(t, strconv.Itoa(*ss.rr.Workers()[0].Pid), res)

	cl, err := rs.Client()
	assert.NoError(t, err)

	r := ""
	assert.NoError(t, cl.Call("http.Reset", true, &r))
	assert.Equal(t, "OK", r)

	res2, _, _ := get("http://localhost:6029")
	assert.Equal(t, strconv.Itoa(*ss.rr.Workers()[0].Pid), res2)
	assert.NotEqual(t, res, res2)
}

func Test_Workers(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		rpcCfg: `{"enable":true, "listen":"tcp://:5004"}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequest": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../php-src/tests/http/client.php pid pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))

	s, _ := c.Get(ID)
	ss := s.(*Service)

	s2, _ := c.Get(rpc.ID)
	rs := s2.(*rpc.Service)

	go func() { c.Serve() }()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	cl, err := rs.Client()
	assert.NoError(t, err)

	r := &WorkerList{}
	assert.NoError(t, cl.Call("http.Workers", true, &r))
	assert.Len(t, r.Workers, 1)

	assert.Equal(t, *ss.rr.Workers()[0].Pid, r.Workers[0].Pid)
}
