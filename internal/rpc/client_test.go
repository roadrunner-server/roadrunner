package rpc_test

import (
	"net"
	"testing"

	"github.com/spiral/roadrunner-binary/v2/internal/rpc"

	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/stretchr/testify/assert"
)

func TestNewClient_RpcServiceDisabled(t *testing.T) {
	cfgPlugin := &config.Plugin{Type: "yaml", ReadInCfg: []byte{}}
	assert.NoError(t, cfgPlugin.Init())

	c, err := rpc.NewClient(cfgPlugin)

	assert.Nil(t, c)
	assert.EqualError(t, err, "rpc service disabled")
}

func TestNewClient_WrongRcpConfiguration(t *testing.T) {
	cfgPlugin := &config.Plugin{Type: "yaml", ReadInCfg: []byte("rpc:\n  $foo bar")}
	assert.NoError(t, cfgPlugin.Init())

	c, err := rpc.NewClient(cfgPlugin)

	assert.Nil(t, c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config_plugin_unmarshal_key")
}

func TestNewClient_ConnectionError(t *testing.T) {
	cfgPlugin := &config.Plugin{Type: "yaml", ReadInCfg: []byte("rpc:\n  listen: tcp://127.0.0.1:0")}
	assert.NoError(t, cfgPlugin.Init())

	c, err := rpc.NewClient(cfgPlugin)

	assert.Nil(t, c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestNewClient_SuccessfullyConnected(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)

	defer func() { assert.NoError(t, l.Close()) }()

	cfgPlugin := &config.Plugin{Type: "yaml", ReadInCfg: []byte("rpc:\n  listen: tcp://" + l.Addr().String())}
	assert.NoError(t, cfgPlugin.Init())

	c, err := rpc.NewClient(cfgPlugin)

	assert.NotNil(t, c)
	assert.NoError(t, err)

	defer func() { assert.NoError(t, c.Close()) }()
}
