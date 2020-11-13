package tests

import (
	"fmt"
	"testing"

	"github.com/spiral/endure"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/metrics"
)

func TestMetricsInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel))
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Viper{}
	cfg.Prefix = "rr"
	cfg.Path = ".rr-test.yaml"

	err = cont.Register(cfg)
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&metrics.Plugin{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&logger.ZapLogger{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	errCh, err := cont.Serve()

	for {
		select {
		case e := <-errCh:
			fmt.Println(e)
		}
	}
}
