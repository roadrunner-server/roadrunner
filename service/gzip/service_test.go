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

// TEST bug #275
func Test_Bug275(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.Error(t, c.Init(&testCfg{
		httpCfg:"",
		gzip: `{"enable":false}`,
	}))

	s, st := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusInactive, st)
}
