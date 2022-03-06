package rpc_test

import (
	"net"
	"testing"

	"github.com/roadrunner-server/roadrunner/v2/internal/rpc"

	"github.com/roadrunner-server/config/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewClient_RpcServiceDisabled(t *testing.T) {
	cfgPlugin := &config.Plugin{Type: "yaml", ReadInCfg: []byte{}}
	assert.NoError(t, cfgPlugin.Init())

	c, err := rpc.NewClient("test/config_rpc_empty.yaml")

	assert.Nil(t, c)
	assert.EqualError(t, err, "rpc service disabled")
}

func TestNewClient_WrongRcpConfiguration(t *testing.T) {
	c, err := rpc.NewClient("test/config_rpc_wrong.yaml")

	assert.Nil(t, c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'' expected a map, got 'string'")
}

func TestNewClient_ConnectionError(t *testing.T) {
	c, err := rpc.NewClient("test/config_rpc_conn_err.yaml")

	assert.Nil(t, c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestNewClient_SuccessfullyConnected(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:55555")
	assert.NoError(t, err)

	defer func() { assert.NoError(t, l.Close()) }()

	c, err := rpc.NewClient("test/config_rpc_ok.yaml")

	assert.NotNil(t, c)
	assert.NoError(t, err)

	defer func() { assert.NoError(t, c.Close()) }()
}
