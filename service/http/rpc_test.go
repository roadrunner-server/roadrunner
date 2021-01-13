package http

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/service/rpc"
	"github.com/stretchr/testify/assert"
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
			"address": ":16031",
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

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("error during the Serve: error %v", err)
		}
	}()

	time.Sleep(time.Second)

	res, _, err := get("http://localhost:16031")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, strconv.Itoa(*ss.rr.Workers()[0].Pid), res)

	cl, err := rs.Client()
	assert.NoError(t, err)

	r := ""
	assert.NoError(t, cl.Call("http.Reset", true, &r))
	assert.Equal(t, "OK", r)

	res2, _, err := get("http://localhost:16031")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, strconv.Itoa(*ss.rr.Workers()[0].Pid), res2)
	assert.NotEqual(t, res, res2)
	c.Stop()
}

func Test_RPC_Unix(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(rpc.ID, &rpc.Service{})
	c.Register(ID, &Service{})

	sock := `unix://` + os.TempDir() + `/rpc.unix`
	data, _ := json.Marshal(sock)

	assert.NoError(t, c.Init(&testCfg{
		rpcCfg: `{"enable":true, "listen":` + string(data) + `}`,
		httpCfg: `{
			"enable": true,
			"address": ":6032",
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

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("error during the Serve: error %v", err)
		}
	}()

	time.Sleep(time.Millisecond * 500)

	res, _, err := get("http://localhost:6032")
	if err != nil {
		c.Stop()
		t.Fatal(err)
	}
	if ss.rr.Workers() != nil && len(ss.rr.Workers()) > 0 {
		assert.Equal(t, strconv.Itoa(*ss.rr.Workers()[0].Pid), res)
	} else {
		c.Stop()
		t.Fatal("no workers initialized")
	}

	cl, err := rs.Client()
	if err != nil {
		c.Stop()
		t.Fatal(err)
	}

	r := ""
	assert.NoError(t, cl.Call("http.Reset", true, &r))
	assert.Equal(t, "OK", r)

	res2, _, err := get("http://localhost:6032")
	if err != nil {
		c.Stop()
		t.Fatal(err)
	}
	assert.Equal(t, strconv.Itoa(*ss.rr.Workers()[0].Pid), res2)
	assert.NotEqual(t, res, res2)
	c.Stop()
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
			"address": ":6033",
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

	go func() {
		err := c.Serve()
		if err != nil {
			t.Errorf("error during the Serve: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 500)

	cl, err := rs.Client()
	assert.NoError(t, err)

	r := &WorkerList{}
	assert.NoError(t, cl.Call("http.Workers", true, &r))
	assert.Len(t, r.Workers, 1)

	assert.Equal(t, *ss.rr.Workers()[0].Pid, r.Workers[0].Pid)
	c.Stop()
}

func Test_Errors(t *testing.T) {
	r := &rpcServer{nil}

	assert.Error(t, r.Reset(true, nil))
	assert.Error(t, r.Workers(true, nil))
}
