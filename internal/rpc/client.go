// Package prc contains wrapper around RPC client ONLY for internal usage.
package rpc

import (
	"net/rpc"

	"github.com/roadrunner-server/config/v2"
	"github.com/roadrunner-server/errors"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	rpcPlugin "github.com/roadrunner-server/rpc/v2"
)

// NewClient creates client ONLY for internal usage (communication between our application with RR side).
// Client will be connected to the RPC.
func NewClient(cfgPlugin *config.Plugin) (*rpc.Client, error) {
	if !cfgPlugin.Has(rpcPlugin.PluginName) {
		return nil, errors.E("rpc service disabled")
	}

	rpcConfig := &rpcPlugin.Config{}
	if err := cfgPlugin.UnmarshalKey(rpcPlugin.PluginName, rpcConfig); err != nil {
		return nil, err
	}

	rpcConfig.InitDefaults()

	conn, err := rpcConfig.Dialer()
	if err != nil {
		return nil, err
	}

	return rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn)), nil
}
