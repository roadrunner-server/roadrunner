// Package prc contains wrapper around RPC client ONLY for internal usage.
package rpc

import (
	"net/rpc"

	"github.com/spiral/errors"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	"github.com/spiral/roadrunner-plugins/v2/config"
	rpcPlugin "github.com/spiral/roadrunner-plugins/v2/rpc"
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
