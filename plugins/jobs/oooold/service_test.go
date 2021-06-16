package oooold

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/service/env"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"syscall"
	"testing"
)

func viperConfig(cfg string) service.Config {
	v := viper.New()
	v.SetConfigType("json")

	err := v.ReadConfig(bytes.NewBuffer([]byte(cfg)))
	if err != nil {
		panic(err)
	}

	return &configWrapper{v}
}

// configWrapper provides interface bridge between v configs and service.Config.
type configWrapper struct {
	v *viper.Viper
}

// Get nested config section (sub-map), returns nil if section not found.
func (w *configWrapper) Get(key string) service.Config {
	sub := w.v.Sub(key)
	if sub == nil {
		return nil
	}

	return &configWrapper{sub}
}

// Unmarshal unmarshal config data into given struct.
func (w *configWrapper) Unmarshal(out interface{}) error {
	return w.v.Unmarshal(out)
}

func jobs(container service.Container) *Service {
	svc, _ := container.Get("jobs")
	return svc.(*Service)
}

func TestService_Init(t *testing.T) {
	c := service.NewContainer(logrus.New())
	c.Register("jobs", &Service{Brokers: map[string]Broker{"ephemeral": &testBroker{}}})

	assert.NoError(t, c.Init(viperConfig(`{
	"jobs":{
		"workers":{
			"command": "php tests/consumer.php",
			"pool.numWorkers": 1
		},
		"pipelines":{"default":{"broker":"ephemeral"}},
    	"dispatch": {
	    	"spiral-jobs-tests-local-*.pipeline": "default"
    	},
    	"consume": ["default"]
	}
}`)))
}

func TestService_ServeStop(t *testing.T) {
	c := service.NewContainer(logrus.New())
	c.Register("env", &env.Service{})
	c.Register("jobs", &Service{Brokers: map[string]Broker{"ephemeral": &testBroker{}}})

	assert.NoError(t, c.Init(viperConfig(`{
	"jobs":{
		"workers":{
			"command": "php tests/consumer.php",
			"pool.numWorkers": 1
		},
		"pipelines":{"default":{"broker":"ephemeral"}},
    	"dispatch": {
	    	"spiral-jobs-tests-local-*.pipeline": "default"
    	},
    	"consume": ["default"]
	}
}`)))

	ready := make(chan interface{})
	jobs(c).AddListener(func(event int, ctx interface{}) {
		if event == EventBrokerReady {
			close(ready)
		}
	})

	go func() { c.Serve() }()
	<-ready
	c.Stop()
}

func TestService_ServeError(t *testing.T) {
	l := logrus.New()
	l.Level = logrus.FatalLevel

	c := service.NewContainer(l)
	c.Register("jobs", &Service{Brokers: map[string]Broker{"ephemeral": &testBroker{}}})

	assert.NoError(t, c.Init(viperConfig(`{
	"jobs":{
		"workers":{
			"command": "php tests/bad-consumer.php",
			"pool.numWorkers": 1
		},
		"pipelines":{"default":{"broker":"ephemeral"}},
    	"dispatch": {
	    	"spiral-jobs-tests-local-*.pipeline": "default"
    	},
    	"consume": ["default"]
	}
}`)))

	assert.Error(t, c.Serve())
}

func TestService_GetPipeline(t *testing.T) {
	c := service.NewContainer(logrus.New())
	c.Register("jobs", &Service{Brokers: map[string]Broker{"ephemeral": &testBroker{}}})

	assert.NoError(t, c.Init(viperConfig(`{
	"jobs":{
		"workers":{
			"command": "php tests/consumer.php",
			"pool.numWorkers": 1
		},
		"pipelines":{"default":{"broker":"ephemeral"}},
    	"dispatch": {
	    	"spiral-jobs-tests-local-*.pipeline": "default"
    	},
    	"consume": ["default"]
	}
}`)))

	ready := make(chan interface{})
	jobs(c).AddListener(func(event int, ctx interface{}) {
		if event == EventBrokerReady {
			close(ready)
		}
	})

	go func() { c.Serve() }()
	defer c.Stop()
	<-ready

	assert.Equal(t, "ephemeral", jobs(c).cfg.pipelines.Get("default").Broker())
}

func TestService_StatPipeline(t *testing.T) {
	c := service.NewContainer(logrus.New())
	c.Register("jobs", &Service{Brokers: map[string]Broker{"ephemeral": &testBroker{}}})

	assert.NoError(t, c.Init(viperConfig(`{
	"jobs":{
		"workers":{
			"command": "php tests/consumer.php",
			"pool.numWorkers": 1
		},
		"pipelines":{"default":{"broker":"ephemeral"}},
    	"dispatch": {
	    	"spiral-jobs-tests-local-*.pipeline": "default"
    	},
    	"consume": ["default"]
	}
}`)))

	ready := make(chan interface{})
	jobs(c).AddListener(func(event int, ctx interface{}) {
		if event == EventBrokerReady {
			close(ready)
		}
	})

	go func() { c.Serve() }()
	defer c.Stop()
	<-ready

	svc := jobs(c)
	pipe := svc.cfg.pipelines.Get("default")

	stat, err := svc.Stat(pipe)
	assert.NoError(t, err)

	assert.Equal(t, int64(0), stat.Queue)
	assert.Equal(t, true, stat.Consuming)
}

func TestService_StatNonConsumingPipeline(t *testing.T) {
	c := service.NewContainer(logrus.New())
	c.Register("jobs", &Service{Brokers: map[string]Broker{"ephemeral": &testBroker{}}})

	assert.NoError(t, c.Init(viperConfig(`{
	"jobs":{
		"workers":{
			"command": "php tests/consumer.php",
			"pool.numWorkers": 1
		},
		"pipelines":{"default":{"broker":"ephemeral"}},
    	"dispatch": {
	    	"spiral-jobs-tests-local-*.pipeline": "default"
    	},
    	"consume": []
	}
}`)))

	ready := make(chan interface{})
	jobs(c).AddListener(func(event int, ctx interface{}) {
		if event == EventBrokerReady {
			close(ready)
		}
	})

	go func() { c.Serve() }()
	defer c.Stop()
	<-ready

	svc := jobs(c)
	pipe := svc.cfg.pipelines.Get("default")

	stat, err := svc.Stat(pipe)
	assert.NoError(t, err)

	assert.Equal(t, int64(0), stat.Queue)
	assert.Equal(t, false, stat.Consuming)
}

func TestService_DoJob(t *testing.T) {
	c := service.NewContainer(logrus.New())
	c.Register("jobs", &Service{Brokers: map[string]Broker{"ephemeral": &testBroker{}}})

	assert.NoError(t, c.Init(viperConfig(`{
	"jobs":{
		"workers":{
			"command": "php tests/consumer.php",
			"pool.numWorkers": 1
		},
		"pipelines":{"default":{"broker":"ephemeral"}},
    	"dispatch": {
	    	"spiral-jobs-tests-local-*.pipeline": "default"
    	},
    	"consume": ["default"]
	}
}`)))

	ready := make(chan interface{})
	jobReady := make(chan interface{})
	jobs(c).AddListener(func(event int, ctx interface{}) {
		if event == EventBrokerReady {
			close(ready)
		}

		if event == EventJobOK {
			close(jobReady)
		}
	})

	go func() { c.Serve() }()
	defer c.Stop()
	<-ready

	svc := jobs(c)

	id, err := svc.Push(&Job{
		Job:     "spiral.jobs.tests.local.job",
		Payload: `{"data":100}`,
		Options: &Options{},
	})
	assert.NoError(t, err)

	<-jobReady

	data, err := ioutil.ReadFile("tests/local.job")
	assert.NoError(t, err)
	defer syscall.Unlink("tests/local.job")

	assert.Contains(t, string(data), id)
}

func TestService_DoUndefinedJob(t *testing.T) {
	c := service.NewContainer(logrus.New())
	c.Register("jobs", &Service{Brokers: map[string]Broker{"ephemeral": &testBroker{}}})

	assert.NoError(t, c.Init(viperConfig(`{
	"jobs":{
		"workers":{
			"command": "php tests/consumer.php",
			"pool.numWorkers": 1
		},
		"pipelines":{"default":{"broker":"ephemeral"}},
    	"dispatch": {
	    	"spiral-jobs-tests-local-*.pipeline": "default"
    	},
    	"consume": ["default"]
	}
}`)))

	ready := make(chan interface{})
	jobs(c).AddListener(func(event int, ctx interface{}) {
		if event == EventBrokerReady {
			close(ready)
		}

	})

	go func() { c.Serve() }()
	defer c.Stop()
	<-ready

	svc := jobs(c)

	_, err := svc.Push(&Job{
		Job:     "spiral.jobs.tests.undefined",
		Payload: `{"data":100}`,
		Options: &Options{},
	})
	assert.Error(t, err)
}

func TestService_DoJobIntoInvalidBroker(t *testing.T) {
	l := logrus.New()
	l.Level = logrus.FatalLevel

	c := service.NewContainer(l)
	c.Register("jobs", &Service{Brokers: map[string]Broker{"ephemeral": &testBroker{}}})

	assert.NoError(t, c.Init(viperConfig(`{
	"jobs":{
		"workers":{
			"command": "php tests/consumer.php",
			"pool.numWorkers": 1
		},
		"pipelines":{"default":{"broker":"undefined"}},
    	"dispatch": {
	    	"spiral-jobs-tests-local-*.pipeline": "default"
    	},
    	"consume": []
	}
}`)))

	ready := make(chan interface{})
	jobs(c).AddListener(func(event int, ctx interface{}) {
		if event == EventBrokerReady {
			close(ready)
		}

	})

	go func() { c.Serve() }()
	defer c.Stop()
	<-ready

	svc := jobs(c)

	_, err := svc.Push(&Job{
		Job:     "spiral.jobs.tests.local.job",
		Payload: `{"data":100}`,
		Options: &Options{},
	})
	assert.Error(t, err)
}

func TestService_DoStatInvalidBroker(t *testing.T) {
	l := logrus.New()
	l.Level = logrus.FatalLevel

	c := service.NewContainer(l)
	c.Register("jobs", &Service{Brokers: map[string]Broker{"ephemeral": &testBroker{}}})

	assert.NoError(t, c.Init(viperConfig(`{
	"jobs":{
		"workers":{
			"command": "php tests/consumer.php",
			"pool.numWorkers": 1
		},
		"pipelines":{"default":{"broker":"undefined"}},
    	"dispatch": {
	    	"spiral-jobs-tests-local-*.pipeline": "default"
    	},
    	"consume": []
	}
}`)))

	ready := make(chan interface{})
	jobs(c).AddListener(func(event int, ctx interface{}) {
		if event == EventBrokerReady {
			close(ready)
		}
	})

	go func() { c.Serve() }()
	defer c.Stop()
	<-ready

	svc := jobs(c)

	_, err := svc.Stat(svc.cfg.pipelines.Get("default"))
	assert.Error(t, err)
}

func TestService_DoErrorJob(t *testing.T) {
	c := service.NewContainer(logrus.New())
	c.Register("jobs", &Service{Brokers: map[string]Broker{"ephemeral": &testBroker{}}})

	assert.NoError(t, c.Init(viperConfig(`{
	"jobs":{
		"workers":{
			"command": "php tests/consumer.php",
			"pool.numWorkers": 1
		},
		"pipelines":{"default":{"broker":"ephemeral"}},
    	"dispatch": {
	    	"spiral-jobs-tests-local-*.pipeline": "default"
    	},
    	"consume": ["default"]
	}
}`)))

	ready := make(chan interface{})
	jobReady := make(chan interface{})

	var jobErr error
	jobs(c).AddListener(func(event int, ctx interface{}) {
		if event == EventBrokerReady {
			close(ready)
		}

		if event == EventJobError {
			jobErr = ctx.(error)
			close(jobReady)
		}
	})

	go func() { c.Serve() }()
	defer c.Stop()
	<-ready

	svc := jobs(c)

	_, err := svc.Push(&Job{
		Job:     "spiral.jobs.tests.local.errorJob",
		Payload: `{"data":100}`,
		Options: &Options{},
	})
	assert.NoError(t, err)

	<-jobReady
	assert.Error(t, jobErr)
	assert.Contains(t, jobErr.Error(), "something is wrong")
}
