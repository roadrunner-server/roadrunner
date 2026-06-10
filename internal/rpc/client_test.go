package rpc_test

import (
	"net"
	"os"
	"testing"

	"github.com/roadrunner-server/config/v6"
	"github.com/roadrunner-server/roadrunner/v2025/internal/rpc"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestNewClient_RpcServiceDisabled(t *testing.T) {
	cfgPlugin := &config.Plugin{Type: "yaml", ReadInCfg: []byte{}}
	assert.NoError(t, cfgPlugin.Init())

	url, c, err := rpc.NewClient("test/config_rpc_empty.yaml", nil)

	assert.Empty(t, url)
	assert.Nil(t, c)
	assert.EqualError(t, err, "rpc service not specified in the configuration. Tip: add\n rpc:\n\r listen: rr_rpc_address")
}

func TestNewClient_WrongRcpConfiguration(t *testing.T) {
	url, c, err := rpc.NewClient("test/config_rpc_wrong.yaml", nil)

	assert.Empty(t, url)
	assert.Nil(t, c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid socket DSN")
}

func TestNewClient_ConnectionError(t *testing.T) {
	url, c, err := rpc.NewClient("test/config_rpc_conn_err.yaml", nil)

	assert.Empty(t, url)
	assert.Nil(t, c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestNewClient_SuccessfullyConnected(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:55554") //nolint:noctx
	assert.NoError(t, err)

	defer func() { assert.NoError(t, l.Close()) }()

	url, c, err := rpc.NewClient("test/config_rpc_ok.yaml", nil)

	assert.Equal(t, "http://127.0.0.1:55554", url)
	assert.NotNil(t, c)
	assert.NoError(t, err)
}

func TestNewClient_WithIncludes(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:6010") //nolint:noctx
	assert.NoError(t, err)

	defer func() { assert.NoError(t, l.Close()) }()

	url, c, err := rpc.NewClient("test/include1/.rr.yaml", nil)

	assert.Equal(t, "http://127.0.0.1:6010", url)
	assert.NotNil(t, c)
	assert.NoError(t, err)
}

func TestNewClient_SuccessfullyConnectedOverride(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:55554") //nolint:noctx
	assert.NoError(t, err)

	defer func() { assert.NoError(t, l.Close()) }()

	url, c, err := rpc.NewClient("test/config_rpc_empty.yaml", []string{"rpc.listen=tcp://127.0.0.1:55554"})

	assert.Equal(t, "http://127.0.0.1:55554", url)
	assert.NotNil(t, c)
	assert.NoError(t, err)
}

// ${} syntax
func TestNewClient_SuccessfullyConnectedEnvDollarSyntax(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:55556") //nolint:noctx
	assert.NoError(t, err)

	defer func() { assert.NoError(t, l.Close()) }()

	require.NoError(t, os.Setenv("RPC", "tcp://127.0.0.1:55556"))
	url, c, err := rpc.NewClient("test/config_rpc_ok_env.yaml", nil)

	assert.Equal(t, "http://127.0.0.1:55556", url)
	assert.NotNil(t, c)
	assert.NoError(t, err)
}
