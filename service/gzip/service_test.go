package gzip

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner/service"
	rrhttp "github.com/spiral/roadrunner/service/http"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testCfg struct {
	gzip    string
	httpCfg string
	//static  string
	target  string
}

func (cfg *testCfg) Get(name string) service.Config {
	if name == rrhttp.ID {
		return &testCfg{target: cfg.httpCfg}
	}

	if name == ID {
		return &testCfg{target: cfg.gzip}
	}
	return nil
}
func (cfg *testCfg) Unmarshal(out interface{}) error {
	return json.Unmarshal([]byte(cfg.target), out)
}

//func get(url string) (string, *http.Response, error) {
//	r, err := http.Get(url)
//	if err != nil {
//		return "", nil, err
//	}
//
//	b, err := ioutil.ReadAll(r.Body)
//	if err != nil {
//		return "", nil, err
//	}
//
//	err = r.Body.Close()
//	if err != nil {
//		return "", nil, err
//	}
//
//	return string(b), r, err
//}

func Test_Disabled(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{
		gzip: `{"enable":false}`,
	}))

	s, st := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusInactive, st)
}

// func Test_Files(t *testing.T) {
// 	logger, _ := test.NewNullLogger()
// 	logger.SetLevel(logrus.DebugLevel)

// 	c := service.NewContainer(logger)
// 	c.Register(rrhttp.ID, &rrhttp.Service{})
// 	c.Register(ID, &Service{})

// 	assert.NoError(t, c.Init(&testCfg{
// 		gzip: `{"enable":true}`,
// 		static: `{"enable":true, "dir":"../../tests", "forbid":[]}`,
// 		httpCfg: `{
// 			"enable": true,
// 			"address": ":6029",
// 			"maxRequestSize": 1024,
// 			"uploads": {
// 				"dir": ` + tmpDir() + `,
// 				"forbid": []
// 			},
// 			"workers":{
// 				"command": "php ../../tests/http/client.php pid pipes",
// 				"relay": "pipes",
// 				"pool": {
// 					"numWorkers": 1,
// 					"allocateTimeout": 10000000,
// 					"destroyTimeout": 10000000
// 				}
// 			}
// 	}`}))

// 	go func() {
// 		err := c.Serve()
// 		if err != nil {
// 			t.Errorf("serve error: %v", err)
// 		}
// 	}()
// 	time.Sleep(time.Millisecond * 1000)
// 	defer c.Stop()

// 	b, _, _ := get("http://localhost:6029/sample.txt")
// 	assert.Equal(t, "sample", b)
// 	//header should not contain content-encoding:gzip because content-length < gziphandler.DefaultMinSize
// 	// b, _, _ := get("http://localhost:6029/gzip-large-file.txt")
// 	//header should contain content-encoding:gzip because content-length > gziphandler.DefaultMinSize
// }

//func tmpDir() string {
//	p := os.TempDir()
//	r, _ := json.Marshal(p)
//
//	return string(r)
//}
