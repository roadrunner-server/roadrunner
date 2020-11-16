package test

import (
	"testing"

	"github.com/spiral/endure"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/stretchr/testify/assert"
)

func TestHTTPInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel), endure.Visualize(endure.StdOut, ""))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   ".rr-http.yaml",
		Prefix: "",
	}
	err = cont.RegisterAll(

		)
	assert.NoError(t, err)
}
