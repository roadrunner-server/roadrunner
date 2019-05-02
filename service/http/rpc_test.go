package http

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/service/rpc"
	"github.com/stretchr/testify/assert"
	"os"
	"runtime"
	"strconv"
	"testing"
	"time"
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
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../tests/http/client.php pid pipes",
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

func Test_RPC_Unix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(ID, &Service{})

	sock := `unix://` + os.TempDir() + `/rpc.unix`
	j, _ := json.Marshal(sock)

	assert.NoError(t, c.Init(&testCfg{
		rpcCfg: `{"enable":true, "listen":` + string(j) + `}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../tests/http/client.php pid pipes",
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
		rpcCfg: `{"enable":true, "listen":"tcp://:5005"}`,
		httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequestSize": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../tests/http/client.php pid pipes",
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

func Test_Errors(t *testing.T) {
	r := &rpcServer{nil}

	assert.Error(t, r.Reset(true, nil))
	assert.Error(t, r.Workers(true, nil))
}
