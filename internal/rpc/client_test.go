package rpc_test

import (
	"net"
	"os"
	"testing"

	"github.com/roadrunner-server/config/v3"
	"github.com/roadrunner-server/roadrunner/v2/internal/rpc"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestNewClient_RpcServiceDisabled(t *testing.T) {
	cfgPlugin := &config.Plugin{Type: "yaml", ReadInCfg: []byte{}}
	assert.NoError(t, cfgPlugin.Init())

	c, err := rpc.NewClient("test/config_rpc_empty.yaml", nil)

	assert.Nil(t, c)
	assert.EqualError(t, err, "rpc service not specified in the configuration. Tip: add\n rpc:\n\r listen: rr_rpc_address")
}

func TestNewClient_WrongRcpConfiguration(t *testing.T) {
	c, err := rpc.NewClient("test/config_rpc_wrong.yaml", nil)

	assert.Nil(t, c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid socket DSN")
}

func TestNewClient_ConnectionError(t *testing.T) {
	c, err := rpc.NewClient("test/config_rpc_conn_err.yaml", nil)

	assert.Nil(t, c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestNewClient_SuccessfullyConnected(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:55555")
	assert.NoError(t, err)

	defer func() { assert.NoError(t, l.Close()) }()

	c, err := rpc.NewClient("test/config_rpc_ok.yaml", nil)

	assert.NotNil(t, c)
	assert.NoError(t, err)

	defer func() { assert.NoError(t, c.Close()) }()
}

func TestNewClient_SuccessfullyConnectedOverride(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:55555")
	assert.NoError(t, err)

	defer func() { assert.NoError(t, l.Close()) }()

	c, err := rpc.NewClient("test/config_rpc_empty.yaml", []string{"rpc.listen=tcp://127.0.0.1:55555"})

	assert.NotNil(t, c)
	assert.NoError(t, err)

	defer func() { assert.NoError(t, c.Close()) }()
}

func TestNewClient_SuccessfullyConnectedEnv(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:55556")
	assert.NoError(t, err)

	defer func() { assert.NoError(t, l.Close()) }()

	require.NoError(t, os.Setenv("RR_RPC_LISTEN", "tcp://127.0.0.1:55556"))
	c, err := rpc.NewClient("test/config_rpc_ok.yaml", nil)

	assert.NotNil(t, c)
	assert.NoError(t, err)

	defer func() { assert.NoError(t, c.Close()) }()
}

// ${} syntax
func TestNewClient_SuccessfullyConnectedEnvDollarSyntax(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:55556")
	assert.NoError(t, err)

	defer func() { assert.NoError(t, l.Close()) }()

	require.NoError(t, os.Setenv("RPC", "tcp://127.0.0.1:55556"))
	c, err := rpc.NewClient("test/config_rpc_ok_env.yaml", nil)

	assert.NotNil(t, c)
	assert.NoError(t, err)

	defer func() { assert.NoError(t, c.Close()) }()
}
